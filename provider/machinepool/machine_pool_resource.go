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

package machinepool

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

var machinepoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9])?$`,
)

type MachinePoolResource struct {
	collection *cmv1.ClustersClient
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
				Description: "Identifier of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Must consist of lower-case alphanumeric characters or '-', start and end with an alphanumeric character.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of machines of the pool",
				Optional:    true,
			},
			"use_spot_instances": schema.BoolAttribute{
				Description: "Use Amazon EC2 Spot Instances.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"max_spot_price": schema.Float64Attribute{
				Description: "Max Spot price.",
				Optional:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
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
				Optional: true,
			},
			"labels": schema.MapAttribute{
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"multi_availability_zone": schema.BoolAttribute{
				Description: "Create a multi-AZ machine pool for a multi-AZ cluster (default is `true`)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Description: "Select the availability zone in which to create a single AZ machine pool for a multi-AZ cluster",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "Select the subnet in which to create a single AZ machine pool for BYO-VPC cluster",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
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
}

func (r *MachinePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan:
	state := &MachinePoolState{}
	diags := req.Plan.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	machinepoolName := state.Name.ValueString()
	if !machinepoolNameRE.MatchString(machinepoolName) {
		resp.Diagnostics.AddError(
			"Cannot create machine pool: ",
			fmt.Sprintf("Cannot create machine pool for cluster '%s' with name '%s'. Expected a valid value for 'name' matching %s",
				state.Cluster.ValueString(), state.Name.ValueString(), machinepoolNameRE,
			),
		)
		return
	}

	// Wait till the cluster is ready:
	err := common.WaitTillClusterReady(ctx, r.collection, state.Cluster.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}

	// Create the machine pool:
	resource := r.collection.Cluster(state.Cluster.ValueString())
	builder := cmv1.NewMachinePool().ID(state.ID.ValueString()).InstanceType(state.MachineType.ValueString())
	builder.ID(state.Name.ValueString())

	if err := setSpotInstances(state, builder); err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s: %v'", state.Cluster.ValueString(), err,
			),
		)
		return
	}

	isMultiAZPool, err := r.validateAZConfig(state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}
	if !common.IsStringAttributeEmpty(state.AvailabilityZone) {
		builder.AvailabilityZones(state.AvailabilityZone.ValueString())
	}
	if !common.IsStringAttributeEmpty(state.SubnetID) {
		builder.Subnets(state.SubnetID.ValueString())
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg := getAutoscaling(state, builder)
	if errMsg != "" {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s, %s'", state.Cluster.ValueString(), errMsg,
			),
		)
		return
	}

	if common.HasValue(state.Replicas) {
		computeNodeEnabled = true
		if isMultiAZPool && state.Replicas.ValueInt64()%3 != 0 {
			resp.Diagnostics.AddError(
				"Cannot build machine pool",
				fmt.Sprintf(
					"Cannot build machine pool for cluster '%s', replicas must be a multiple of 3",
					state.Cluster.ValueString(),
				),
			)
			return
		}
		builder.Replicas(int(state.Replicas.ValueInt64()))
	}
	if (!autoscalingEnabled && !computeNodeEnabled) || (autoscalingEnabled && computeNodeEnabled) {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s', please provide a value for either the 'replicas' or 'autoscaling_enabled' parameter. It is mandatory to include at least one of these parameters in the resource plan.",
				state.Cluster.ValueString(),
			),
		)
		return
	}

	if state.Taints != nil && len(state.Taints) > 0 {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range state.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().
				Key(taint.Key.ValueString()).
				Value(taint.Value.ValueString()).
				Effect(taint.ScheduleType.ValueString()))
		}
		builder.Taints(taintBuilders...)
	}

	if common.HasValue(state.Labels) {
		labels := map[string]string{}
		for k, v := range state.Labels.Elements() {
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
				state.Cluster.ValueString(), err,
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
				state.Cluster.ValueString(), err,
			),
		)
		return
	}
	object = add.Body()

	// Save the state:
	r.populateState(object, state)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *MachinePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Find the machine pool:
	resource := r.collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString())
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("machine pool (%s) of cluster (%s) not found, removing from state",
			state.ID.ValueString(), state.Cluster.ValueString(),
		))
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to fetch machine pool",
			fmt.Sprintf(
				"Failed to fetch machine pool with identifier %s for cluster %s. Response code: %v",
				state.ID.ValueString(), state.Cluster.ValueString(), get.Status(),
			),
		)
		return
	}

	object := get.Body()

	// Save the state:
	r.populateState(object, state)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
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

	resource := r.collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString())
	_, err := resource.Get().SendContext(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot find machine pool",
			fmt.Sprintf(
				"Cannot find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return
	}

	mpBuilder := cmv1.NewMachinePool().ID(state.ID.ValueString())

	_, ok := common.ShouldPatchString(state.MachineType, plan.MachineType)
	if ok {
		resp.Diagnostics.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s', machine type cannot be updated",
				state.Cluster.ValueString(),
			),
		)
		return
	}

	computeNodesEnabled := false
	autoscalingEnabled := false

	if common.HasValue(plan.Replicas) {
		computeNodesEnabled = true
		mpBuilder.Replicas(int(plan.Replicas.ValueInt64()))

	}

	autoscalingEnabled, errMsg := getAutoscaling(plan, mpBuilder)
	if errMsg != "" {
		resp.Diagnostics.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s, %s ", state.Cluster.ValueString(), errMsg,
			),
		)
		return
	}

	if (autoscalingEnabled && computeNodesEnabled) || (!autoscalingEnabled && !computeNodesEnabled) {
		resp.Diagnostics.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: either autoscaling or compute nodes should be enabled", state.Cluster.ValueString(),
			),
		)
		return
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
		resp.Diagnostics.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: %v ", state.Cluster.ValueString(), err,
			),
		)
		return
	}
	update, err := r.collection.Cluster(state.Cluster.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString()).Update().Body(machinePool).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update machine pool",
			fmt.Sprintf(
				"Failed to update machine pool '%s'  on cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return
	}

	object := update.Body()

	// Save the state:
	r.populateState(object, plan)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
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
	clusterSubnets := cluster.AWS().SubnetIDs()

	if isMultiAZCluster {
		// Can't set both availability_zone and subnet_id
		if !common.IsStringAttributeEmpty(state.AvailabilityZone) && !common.IsStringAttributeEmpty(state.SubnetID) {
			return false, fmt.Errorf("availability_zone and subnet_id are mutually exclusive")
		}

		// multi_availability_zone setting must be consistent with availability_zone and subnet_id
		azOrSubnet := !common.IsStringAttributeEmpty(state.AvailabilityZone) || !common.IsStringAttributeEmpty(state.SubnetID)
		if common.HasValue(state.MultiAvailabilityZone) {
			if azOrSubnet && state.MultiAvailabilityZone.ValueBool() {
				return false, fmt.Errorf("multi_availability_zone must be False when availability_zone or subnet_id is set")
			}
		} else {
			state.MultiAvailabilityZone = types.BoolValue(!azOrSubnet)
		}
	} else { // not a multi-AZ cluster
		if !common.IsStringAttributeEmpty(state.AvailabilityZone) {
			return false, fmt.Errorf("availability_zone can only be set for multi-AZ clusters")
		}
		if common.HasValue(state.MultiAvailabilityZone) && state.MultiAvailabilityZone.ValueBool() {
			return false, fmt.Errorf("multi_availability_zone can only be set for multi-AZ clusters")
		}
		state.MultiAvailabilityZone = types.BoolValue(false)
	}

	// Ensure that the machine pool's AZ and subnet are valid for the cluster
	// If subnet is set, we make sure it's valid for the cluster, but we don't default it if not set
	if !common.IsStringAttributeEmpty(state.SubnetID) {
		inClusterSubnet := false
		for _, subnet := range clusterSubnets {
			if subnet == state.SubnetID.ValueString() {
				inClusterSubnet = true
				break
			}
		}
		if !inClusterSubnet {
			return false, fmt.Errorf("subnet_id %s is not valid for cluster %s", state.SubnetID.ValueString(), state.Cluster.ValueString())
		}
	} else {
		state.SubnetID = types.StringNull()
	}
	// If AZ is set, we make sure it's valid for the cluster. If not set and neither is subnet, we default it to the 1st AZ in the cluster
	if !common.IsStringAttributeEmpty(state.AvailabilityZone) {
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
		if len(clusterAZs) > 0 && !state.MultiAvailabilityZone.ValueBool() && isMultiAZCluster && common.IsStringAttributeEmpty(state.SubnetID) {
			state.AvailabilityZone = types.StringValue(clusterAZs[0])
		} else {
			state.AvailabilityZone = types.StringNull()
		}
	}
	return state.MultiAvailabilityZone.ValueBool(), nil
}

func setSpotInstances(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) error {
	useSpotInstances := common.HasValue(state.UseSpotInstances) && state.UseSpotInstances.ValueBool()
	isSpotMaxPriceSet := common.HasValue(state.MaxSpotPrice)

	if isSpotMaxPriceSet && !useSpotInstances {
		return errors.New("Cannot set max price when not using spot instances (set \"use_spot_instances\" to true)")
	}

	if useSpotInstances {
		awsMachinePool := cmv1.NewAWSMachinePool()
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions()
		if isSpotMaxPriceSet {
			spotMarketOptions.MaxPrice(state.MaxSpotPrice.ValueFloat64())
		}
		awsMachinePool.SpotMarketOptions(spotMarketOptions)
		mpBuilder.AWS(awsMachinePool)
	}

	return nil
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

	// Remove the state:
	resp.State.RemoveResource(ctx)
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
func (r *MachinePoolResource) populateState(object *cmv1.MachinePool, state *MachinePoolState) {
	state.ID = types.StringValue(object.ID())
	state.Name = types.StringValue(object.ID())

	if getAWS, ok := object.GetAWS(); ok {
		if spotMarketOptions, ok := getAWS.GetSpotMarketOptions(); ok {
			state.UseSpotInstances = types.BoolValue(true)
			if spotMarketOptions.MaxPrice() != 0 {
				state.MaxSpotPrice = types.Float64Value(spotMarketOptions.MaxPrice())
			}
		}
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
	azs := object.AvailabilityZones()
	if len(azs) == 1 {
		state.AvailabilityZone = types.StringValue(azs[0])
	} else {
		state.AvailabilityZone = types.StringValue("")
	}

	subnets := object.Subnets()
	if len(subnets) == 1 {
		state.SubnetID = types.StringValue(subnets[0])
	} else {
		state.SubnetID = types.StringValue("")
	}
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
