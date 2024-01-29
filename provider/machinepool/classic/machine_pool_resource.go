/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package classic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// This is a magic name to trigger special handling for the cluster's default
// machine pool
const defaultMachinePoolName = "worker"

var machinepoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9])?$`,
)

type MachinePoolResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

var _ resource.ResourceWithConfigure = &MachinePoolResource{}
var _ resource.ResourceWithImportState = &MachinePoolResource{}
var _ resource.ResourceWithConfigValidators = &MachinePoolResource{}

func New() resource.Resource {
	return &MachinePoolResource{}
}

func (r *MachinePoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_pool"
}

func (r *MachinePoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Machine pool.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Must consist of lower-case alphanumeric characters or '-', start and end with an alphanumeric character. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values. " + common.ValueCannotBeChangedStringDescription,
				Required: true,
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of machines of the pool",
				Optional:    true,
			},
			"use_spot_instances": schema.BoolAttribute{
				Description: "Use Amazon EC2 Spot Instances. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
			},
			"max_spot_price": schema.Float64Attribute{
				Description: "Max Spot price. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Validators: []validator.Float64{
					float64validator.AtLeast(1e-6), // Greater than zero
				},
			},
			"autoscaling_enabled": schema.BoolAttribute{
				Description: "Enables autoscaling. If `true`, this variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
				Optional:    true,
			},
			"min_replicas": schema.Int64Attribute{
				Description: "The minimum number of replicas for autoscaling functionality.",
				Optional:    true,
			},
			"max_replicas": schema.Int64Attribute{
				Description: "The maximum number of replicas for autoscaling functionality.",
				Optional:    true,
			},
			"taints": schema.ListNestedAttribute{
				Description: "Taints for a machine pool. Format should be a comma-separated " +
					"list of 'key=value'. This list will overwrite any modifications " +
					"made to node taints on an ongoing basis.\n",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Taints key",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Taints value",
							Required:    true,
						},
						"schedule_type": schema.StringAttribute{
							Description: "Taints schedule type",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("NoSchedule", "PreferNoSchedule", "NoExecute"),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				Optional: true,
			},
			"labels": schema.MapAttribute{
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
			},
			"multi_availability_zone": schema.BoolAttribute{
				Description: "Create a multi-AZ machine pool for a multi-AZ cluster (default is `true`). " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Description: "Select the availability zone in which to create a single AZ machine pool for a multi-AZ cluster. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zones": schema.ListAttribute{
				Description: "A list of Availability Zones. Relevant only for multiple availability zones machine pool. For single availability zone check \"availability_zone\" attribute.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"subnet_id": schema.StringAttribute{
				Description: "Select the subnet in which to create a single AZ machine pool for BYO-VPC cluster. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_ids": schema.ListAttribute{
				Description: "A list of IDs of subnets in which the machines of this machine pool are created. Relevant only for a machine pool with multiple subnets. For machine pool with single subnet check \"subnet_id\" attribute",
				ElementType: types.StringType,
				Computed:    true,
			},
			"disk_size": schema.Int64Attribute{
				Description: "Root disk size, in GiB. " + common.ValueCannotBeChangedStringDescription,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"aws_additional_security_group_ids": schema.ListAttribute{
				Description: "AWS additional security group ids. " + common.ValueCannotBeChangedStringDescription,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *MachinePoolResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(path.MatchRoot("availability_zone"), path.MatchRoot("subnet_id")),
		resourcevalidator.RequiredTogether(path.MatchRoot("min_replicas"), path.MatchRoot("max_replicas")),
		resourcevalidator.Conflicting(path.MatchRoot("replicas"), path.MatchRoot("min_replicas")),
		resourcevalidator.Conflicting(path.MatchRoot("replicas"), path.MatchRoot("max_replicas")),
	}
}

func (r *MachinePoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = connection.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection)
}

func (r *MachinePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan:
	plan := &MachinePoolState{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	machinepoolName := plan.Name.ValueString()
	if !machinepoolNameRE.MatchString(machinepoolName) {
		resp.Diagnostics.AddError(
			"Cannot create machine pool: ",
			fmt.Sprintf("Cannot create machine pool for cluster '%s' with name '%s'. Expected a valid value for 'name' matching %s",
				plan.Cluster.ValueString(), plan.Name.ValueString(), machinepoolNameRE,
			),
		)
		return
	}

	// Wait till the cluster is ready:
	err := r.clusterWait.WaitForClusterToBeReady(ctx, plan.Cluster.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	// The default machine pool is created automatically when the cluster is created.
	// We want to import it instead of creating it.
	if machinepoolName == defaultMachinePoolName {
		r.magicImport(ctx, plan, resp)
		return
	}

	// Create the machine pool:
	resource := r.collection.Cluster(plan.Cluster.ValueString())
	builder := cmv1.NewMachinePool().ID(plan.ID.ValueString()).InstanceType(plan.MachineType.ValueString())
	builder.ID(plan.Name.ValueString())

	if workerDiskSize := common.OptionalInt64(plan.DiskSize); workerDiskSize != nil {
		builder.RootVolume(
			cmv1.NewRootVolume().AWS(
				cmv1.NewAWSVolume().Size(int(*workerDiskSize)),
			),
		)
	}

	awsMachinePoolBuilder, err := setSpotInstances(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s: %v'", plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	isMultiAZPool, err := r.validateAZConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}
	if !common.IsStringAttributeUnknownOrEmpty(plan.AvailabilityZone) {
		builder.AvailabilityZones(plan.AvailabilityZone.ValueString())
	}
	if !common.IsStringAttributeUnknownOrEmpty(plan.SubnetID) {
		builder.Subnets(plan.SubnetID.ValueString())
	}
	if common.HasValue(plan.AdditionalSecurityGroupIds) {
		if awsMachinePoolBuilder == nil {
			awsMachinePoolBuilder = cmv1.NewAWSMachinePool()
		}
		additionalSecurityGroupIds, err := common.StringListToArray(ctx, plan.AdditionalSecurityGroupIds)
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot convert Additional Security Groups to slice",
				fmt.Sprintf(
					"Cannot convert Additional Security Groups to slice for cluster '%s: %v'", plan.Cluster.ValueString(), err,
				),
			)
			return
		}
		awsMachinePoolBuilder.AdditionalSecurityGroupIds(additionalSecurityGroupIds...)
	}
	if awsMachinePoolBuilder != nil {
		builder.AWS(awsMachinePoolBuilder)
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg := getAutoscaling(plan, builder)
	if errMsg != "" {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s, %s'", plan.Cluster.ValueString(), errMsg,
			),
		)
		return
	}

	if common.HasValue(plan.Replicas) {
		computeNodeEnabled = true
		if isMultiAZPool && plan.Replicas.ValueInt64()%3 != 0 {
			resp.Diagnostics.AddError(
				"Cannot build machine pool",
				fmt.Sprintf(
					"Cannot build machine pool for cluster '%s', replicas must be a multiple of 3",
					plan.Cluster.ValueString(),
				),
			)
			return
		}
		builder.Replicas(int(plan.Replicas.ValueInt64()))
	}
	if (!autoscalingEnabled && !computeNodeEnabled) || (autoscalingEnabled && computeNodeEnabled) {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s', please provide a value for either the 'replicas' or 'autoscaling_enabled' parameter. It is mandatory to include at least one of these parameters in the resource plan.",
				plan.Cluster.ValueString(),
			),
		)
		return
	}

	if plan.Taints != nil && len(plan.Taints) > 0 {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range plan.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().
				Key(taint.Key.ValueString()).
				Value(taint.Value.ValueString()).
				Effect(taint.ScheduleType.ValueString()))
		}
		builder.Taints(taintBuilders...)
	}

	if common.HasValue(plan.Labels) {
		labels := map[string]string{}
		for k, v := range plan.Labels.Elements() {
			labels[k] = v.(types.String).ValueString()
		}
		builder.Labels(labels)
	}

	object, err := builder.Build()
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	collection := resource.MachinePools()
	add, err := collection.Add().Body(object).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create machine pool",
			fmt.Sprintf(
				"Cannot create machine pool for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}
	object = add.Body()

	// Save the state:
	err = populateState(object, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't populate machine pool state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// This handles the "magic" import of the default machine pool, allowing the
// user to include it in their config w/o having to specifically `terraform
// import` it.
func (r *MachinePoolResource) magicImport(ctx context.Context, plan *MachinePoolState, resp *resource.CreateResponse) {
	machinepoolName := plan.Name.ValueString()
	state := &MachinePoolState{
		ID:      types.StringValue(machinepoolName),
		Cluster: plan.Cluster,
		Name:    types.StringValue(machinepoolName),
	}
	plan.ID = types.StringValue(machinepoolName)

	notFound, diags := readState(ctx, state, r.collection)
	if notFound {
		// We disallow creating a machine pool with the default name. This
		// case can only happen if the default machine pool was deleted and
		// the user tries to recreate it.
		diags.AddError(
			"Can't create machine pool",
			fmt.Sprintf(
				"Can't create machine pool for cluster '%s': "+
					"the default machine pool '%s' was deleted and a new machine pool with that name may not be created. "+
					"Please use a different name.",
				plan.Cluster.ValueString(),
				machinepoolName,
			),
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = r.doUpdate(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *MachinePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notFound, diags := readState(ctx, state, r.collection)
	if notFound {
		// If we can't find the machine pool, it was deleted. Remove if from the
		// state and don't return an error so the TF apply() will automatically
		// recreate it.
		tflog.Warn(ctx, fmt.Sprintf("machine pool (%s) of cluster (%s) not found, removing from state",
			state.ID.ValueString(), state.Cluster.ValueString(),
		))
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func readState(ctx context.Context, state *MachinePoolState, collection *cmv1.ClustersClient) (poolNotFound bool, diags diag.Diagnostics) {
	diags = diag.Diagnostics{}

	resource := collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString())
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		poolNotFound = true
		return
	} else if err != nil {
		diags.AddError(
			"Failed to fetch machine pool",
			fmt.Sprintf(
				"Failed to fetch machine pool with identifier %s for cluster %s. Response code: %v",
				state.ID.ValueString(), state.Cluster.ValueString(), get.Status(),
			),
		)
		return
	}

	object := get.Body()
	err = populateState(object, state)
	if err != nil {
		diags.AddError(
			"Can't populate machine pool state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	return
}

func validateNoImmutableAttChange(state, plan *MachinePoolState) diag.Diagnostics {
	diags := diag.Diagnostics{}
	validateStateAndPlanEquals(state.Cluster, plan.Cluster, "cluster", &diags)
	validateStateAndPlanEquals(state.Name, plan.Name, "name", &diags)
	validateStateAndPlanEquals(state.MachineType, plan.MachineType, "machine_type", &diags)
	validateStateAndPlanEquals(state.UseSpotInstances, plan.UseSpotInstances, "use_spot_instances", &diags)
	validateStateAndPlanEquals(state.MaxSpotPrice, plan.MaxSpotPrice, "max_spot_price", &diags)
	validateStateAndPlanEquals(state.MultiAvailabilityZone, plan.MultiAvailabilityZone, "multi_availability_zone", &diags)
	validateStateAndPlanEquals(state.AvailabilityZone, plan.AvailabilityZone, "availability_zone", &diags)
	validateStateAndPlanEquals(state.SubnetID, plan.SubnetID, "subnet_id", &diags)
	validateStateAndPlanEquals(state.DiskSize, plan.DiskSize, "disk_size", &diags)
	validateStateAndPlanEquals(state.AdditionalSecurityGroupIds, plan.AdditionalSecurityGroupIds, "aws_additional_security_group_ids", &diags)

	return diags

}

func validateStateAndPlanEquals(stateAttr attr.Value, planAttr attr.Value, attrName string, diags *diag.Diagnostics) {
	// Its possible to have here unknown attributes
	// Relevant only for optional computed attributes in resource create
	// Check this because this function also used in "magicImport" function
	if planAttr.IsUnknown() {
		return
	}
	common.ValidateStateAndPlanEquals(stateAttr, planAttr, attrName, diags)
}

func (r *MachinePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get the state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &MachinePoolState{}
	diags = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.doUpdate(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *MachinePoolResource) doUpdate(ctx context.Context, state *MachinePoolState, plan *MachinePoolState) diag.Diagnostics {
	//assert no changes on specific attributes
	diags := validateNoImmutableAttChange(state, plan)
	if diags.HasError() {
		return diags
	}

	resource := r.collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString())
	_, err := resource.Get().SendContext(ctx)

	if err != nil {
		diags.AddError(
			"Cannot find machine pool",
			fmt.Sprintf(
				"Cannot find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return diags
	}

	mpBuilder := cmv1.NewMachinePool().ID(state.ID.ValueString())

	_, ok := common.ShouldPatchString(state.MachineType, plan.MachineType)
	if ok {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s', machine type cannot be updated",
				state.Cluster.ValueString(),
			),
		)
		return diags
	}

	computeNodesEnabled := false
	autoscalingEnabled := false

	if common.HasValue(plan.Replicas) {
		computeNodesEnabled = true
		mpBuilder.Replicas(int(plan.Replicas.ValueInt64()))

	}

	autoscalingEnabled, errMsg := getAutoscaling(plan, mpBuilder)
	if errMsg != "" {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s, %s ", state.Cluster.ValueString(), errMsg,
			),
		)
		return diags
	}

	if (autoscalingEnabled && computeNodesEnabled) || (!autoscalingEnabled && !computeNodesEnabled) {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: either replicas should be set or autoscaling enabled", state.Cluster.ValueString(),
			),
		)
		return diags
	}

	patchLabels, shouldPatchLabels := common.ShouldPatchMap(state.Labels, plan.Labels)
	if shouldPatchLabels {
		labels := map[string]string{}
		for k, v := range patchLabels.Elements() {
			labels[k] = v.(types.String).ValueString()
		}
		mpBuilder.Labels(labels)
	}

	if shouldPatchTaints(state.Taints, plan.Taints) {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range plan.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().
				Key(taint.Key.ValueString()).
				Value(taint.Value.ValueString()).
				Effect(taint.ScheduleType.ValueString()))
		}
		mpBuilder.Taints(taintBuilders...)
	}

	machinePool, err := mpBuilder.Build()
	if err != nil {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: %v ", state.Cluster.ValueString(), err,
			),
		)
		return diags
	}
	update, err := r.collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString()).Update().Body(machinePool).SendContext(ctx)
	if err != nil {
		diags.AddError(
			"Failed to update machine pool",
			fmt.Sprintf(
				"Failed to update machine pool '%s'  on cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return diags
	}

	object := update.Body()

	// update the autoscaling enabled with the plan value (important for nil and false cases)
	state.AutoScalingEnabled = plan.AutoScalingEnabled
	// update the Replicas with the plan value (important for nil and zero value cases)
	state.Replicas = plan.Replicas

	// Save the state:
	err = populateState(object, state)
	if err != nil {
		diags.AddError(
			"Can't populate machine pool state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return diags
	}
	return diags
}

// Validate the machine pool's settings that pertain to availability zones.
// Returns whether the machine pool is/will be multi-AZ.
func (r *MachinePoolResource) validateAZConfig(state *MachinePoolState) (bool, error) {
	resp, err := r.collection.Cluster(state.Cluster.ValueString()).Get().Send()
	if err != nil {
		return false, fmt.Errorf("failed to get information for cluster %s: %v", state.Cluster.ValueString(), err)
	}
	cluster := resp.Body()
	isMultiAZCluster := cluster.MultiAZ()
	clusterAZs := cluster.Nodes().AvailabilityZones()

	if isMultiAZCluster {
		// Can't set both availability_zone and subnet_id
		if !common.IsStringAttributeUnknownOrEmpty(state.AvailabilityZone) && !common.IsStringAttributeUnknownOrEmpty(state.SubnetID) {
			return false, fmt.Errorf("availability_zone and subnet_id are mutually exclusive")
		}

		// multi_availability_zone setting must be consistent with availability_zone and subnet_id
		azOrSubnet := !common.IsStringAttributeUnknownOrEmpty(state.AvailabilityZone) || !common.IsStringAttributeUnknownOrEmpty(state.SubnetID)
		if common.HasValue(state.MultiAvailabilityZone) {
			if azOrSubnet && state.MultiAvailabilityZone.ValueBool() {
				return false, fmt.Errorf("multi_availability_zone must be False when availability_zone or subnet_id is set")
			}
		} else {
			state.MultiAvailabilityZone = types.BoolValue(!azOrSubnet)
		}
	} else { // not a multi-AZ cluster
		if !common.IsStringAttributeUnknownOrEmpty(state.AvailabilityZone) {
			return false, fmt.Errorf("availability_zone can only be set for multi-AZ clusters")
		}
		if common.HasValue(state.MultiAvailabilityZone) && state.MultiAvailabilityZone.ValueBool() {
			return false, fmt.Errorf("multi_availability_zone can only be set for multi-AZ clusters")
		}
		state.MultiAvailabilityZone = types.BoolValue(false)
	}

	// If AZ is set, we make sure it's valid for the cluster. If not set and neither is subnet, we default it to the 1st AZ in the cluster
	if !common.IsStringAttributeUnknownOrEmpty(state.AvailabilityZone) {
		inClusterAZ := false
		for _, az := range clusterAZs {
			if az == state.AvailabilityZone.ValueString() {
				inClusterAZ = true
				break
			}
		}
		if !inClusterAZ {
			return false, fmt.Errorf("availability_zone %s is not valid for cluster %s", state.AvailabilityZone.ValueString(), state.Cluster.ValueString())
		}
	} else {
		if len(clusterAZs) > 0 && !state.MultiAvailabilityZone.ValueBool() && isMultiAZCluster && common.IsStringAttributeUnknownOrEmpty(state.SubnetID) {
			state.AvailabilityZone = types.StringValue(clusterAZs[0])
		} else {
			state.AvailabilityZone = types.StringNull()
		}
	}
	return state.MultiAvailabilityZone.ValueBool(), nil
}

func setSpotInstances(state *MachinePoolState) (*cmv1.AWSMachinePoolBuilder, error) {
	useSpotInstances := common.HasValue(state.UseSpotInstances) && state.UseSpotInstances.ValueBool()
	isSpotMaxPriceSet := common.HasValue(state.MaxSpotPrice)
	var awsMachinePool *cmv1.AWSMachinePoolBuilder

	if isSpotMaxPriceSet && !useSpotInstances {
		return awsMachinePool, errors.New("Cannot set max price when not using spot instances (set \"use_spot_instances\" to true)")
	}

	if useSpotInstances {
		awsMachinePool = cmv1.NewAWSMachinePool()
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions()
		if isSpotMaxPriceSet {
			spotMarketOptions.MaxPrice(state.MaxSpotPrice.ValueFloat64())
		}
		awsMachinePool.SpotMarketOptions(spotMarketOptions)
	}

	return awsMachinePool, nil
}

func getAutoscaling(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) (
	autoscalingEnabled bool, errMsg string) {
	autoscalingEnabled = false
	if common.HasValue(state.AutoScalingEnabled) && state.AutoScalingEnabled.ValueBool() {
		autoscalingEnabled = true

		autoscaling := cmv1.NewMachinePoolAutoscaling()
		if common.HasValue(state.MaxReplicas) {
			autoscaling.MaxReplicas(int(state.MaxReplicas.ValueInt64()))
		} else {
			return false, "when enabling autoscaling, should set value for maxReplicas"
		}
		if common.HasValue(state.MinReplicas) {
			autoscaling.MinReplicas(int(state.MinReplicas.ValueInt64()))
		} else {
			return false, "when enabling autoscaling, should set value for minReplicas"
		}
		if !autoscaling.Empty() {
			mpBuilder.Autoscaling(autoscaling)
		}
	} else {
		if common.HasValue(state.MaxReplicas) || common.HasValue(state.MinReplicas) {
			return false, "when disabling autoscaling, cannot set min_replicas and/or max_replicas"
		}
	}

	return autoscalingEnabled, ""
}

func (r *MachinePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the machine pool:
	resource := r.collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		// We can't delete the pool, see if it's the last one:
		numPools, err2 := r.countPools(ctx, state.Cluster.ValueString())
		if numPools == 1 && err2 == nil {
			// It's the last one, issue warning instead of error
			resp.Diagnostics.AddWarning(
				"Cannot delete machine pool",
				fmt.Sprintf(
					"Cannot delete the last machine pool for cluster '%s'. "+
						"ROSA Classic clusters must have at least one machine pool. "+
						"It is being removed from the Terraform state only. "+
						"To resume managing this machine pool, import it again. "+
						"It will be automatically deleted when the cluster is deleted.",
					state.Cluster.ValueString(),
				),
			)
			// No return, we want to remove the state
		} else {
			// Wasn't the last one, return error
			resp.Diagnostics.AddError(
				"Cannot delete machine pool",
				fmt.Sprintf(
					"Cannot delete machine pool with identifier '%s' for "+
						"cluster '%s': %v",
					state.ID.ValueString(), state.Cluster.ValueString(), err,
				),
			)
			return
		}
	}

	// Remove the state:
	resp.State.RemoveResource(ctx)
}

// countPools returns the number of machine pools in the given cluster
func (r *MachinePoolResource) countPools(ctx context.Context, clusterID string) (int, error) {
	resource := r.collection.Cluster(clusterID).MachinePools()
	resp, err := resource.List().SendContext(ctx)
	if err != nil {
		return 0, err
	}
	return resp.Size(), nil
}

func (r *MachinePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// To import a machine pool, we need to know the cluster ID and the machine pool ID
	fields := strings.Split(req.ID, ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Machine pool to import should be specified as <cluster_id>,<machine_pool_id>",
		)
		return
	}
	clusterID := fields[0]
	machinePoolID := fields[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), machinePoolID)...)
}

// populateState copies the data from the API object to the Terraform state.
func populateState(object *cmv1.MachinePool, state *MachinePoolState) error {
	state.ID = types.StringValue(object.ID())
	state.Name = types.StringValue(object.ID())

	if getAWS, ok := object.GetAWS(); ok {
		if spotMarketOptions, ok := getAWS.GetSpotMarketOptions(); ok {
			state.UseSpotInstances = types.BoolValue(true)
			if spotMarketOptions.MaxPrice() != 0 {
				state.MaxSpotPrice = types.Float64Value(spotMarketOptions.MaxPrice())
			}
		} else {
			state.UseSpotInstances = types.BoolNull()
		}
		if additionalSecurityGroups, ok := getAWS.GetAdditionalSecurityGroupIds(); ok {
			additionalSecurityGroupsList, err := common.StringArrayToList(additionalSecurityGroups)
			if err != nil {
				return err
			}
			state.AdditionalSecurityGroupIds = additionalSecurityGroupsList
		} else {
			state.AdditionalSecurityGroupIds = types.ListNull(types.StringType)
		}
	} else {
		state.AdditionalSecurityGroupIds = types.ListNull(types.StringType)
	}

	autoscaling, ok := object.GetAutoscaling()
	if ok {
		var minReplicas, maxReplicas int
		state.AutoScalingEnabled = types.BoolValue(true)
		minReplicas, ok = autoscaling.GetMinReplicas()
		if ok {
			state.MinReplicas = types.Int64Value(int64(minReplicas))
		}
		maxReplicas, ok = autoscaling.GetMaxReplicas()
		if ok {
			state.MaxReplicas = types.Int64Value(int64(maxReplicas))
		}
	} else {
		state.MaxReplicas = types.Int64Null()
		state.MinReplicas = types.Int64Null()
	}

	if instanceType, ok := object.GetInstanceType(); ok {
		state.MachineType = types.StringValue(instanceType)
	}

	if replicas, ok := object.GetReplicas(); ok {
		state.Replicas = types.Int64Value(int64(replicas))
	}

	taints := object.Taints()
	if len(taints) > 0 {
		state.Taints = make([]Taints, len(taints))
		for i, taint := range taints {
			state.Taints[i] = Taints{
				Key:          types.StringValue(taint.Key()),
				Value:        types.StringValue(taint.Value()),
				ScheduleType: types.StringValue(taint.Effect()),
			}
		}
	} else {
		state.Taints = nil
	}

	labels := object.Labels()
	if len(labels) > 0 {
		// XXX: We should be checking error here, but we don't have a way to return the error
		state.Labels, _ = common.ConvertStringMapToMapType(labels)
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	// Due to RequiresReplace(), we need to ensure these fields always have a
	// value, even if it's empty. It will be empty if the cluster is multi-AZ or
	// if the other (AZ/subnet) value is set. We don't need to set
	// MultiAvailibilityZone here, because it's set in the validation function
	// during create.
	state.MultiAvailabilityZone = types.BoolValue(true)
	azs := object.AvailabilityZones()
	if len(azs) == 1 {
		state.AvailabilityZone = types.StringValue(azs[0])
		state.MultiAvailabilityZone = types.BoolValue(false)
		state.AvailabilityZones = types.ListNull(types.StringType)
	} else {
		state.AvailabilityZone = types.StringValue("")
		azListValue, err := common.StringArrayToList(azs)
		if err != nil {
			return err
		}
		state.AvailabilityZones = azListValue
	}

	subnets := object.Subnets()
	if len(subnets) == 1 {
		state.SubnetID = types.StringValue(subnets[0])
		state.MultiAvailabilityZone = types.BoolValue(false)
		state.SubnetIDs = types.ListNull(types.StringType)
	} else {
		state.SubnetID = types.StringValue("")
		subnetListValue, err := common.StringArrayToList(subnets)
		if err != nil {
			return err
		}
		state.SubnetIDs = subnetListValue
	}

	state.DiskSize = types.Int64Null()
	if rv, ok := object.GetRootVolume(); ok {
		if aws, ok := rv.GetAWS(); ok {
			if workerDiskSize, ok := aws.GetSize(); ok {
				state.DiskSize = types.Int64Value(int64(workerDiskSize))
			}
		}
	}
	return nil
}

func shouldPatchTaints(a, b []Taints) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return true
	}
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if !a[i].Key.Equal(b[i].Key) || !a[i].Value.Equal(b[i].Value) || !a[i].ScheduleType.Equal(b[i].ScheduleType) {
			return true
		}
	}
	return false
}
