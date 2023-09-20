/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"context"
	"errors"
***REMOVED***
***REMOVED***
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

// This is a magic name to trigger special handling for the cluster's default
// machine pool
const defaultMachinePoolName = "worker"

var machinepoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9]***REMOVED***?$`,
***REMOVED***

type MachinePoolResource struct {
	collection *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &MachinePoolResource{}
var _ resource.ResourceWithImportState = &MachinePoolResource{}
var _ resource.ResourceWithConfigValidators = &MachinePoolResource{}

func New(***REMOVED*** resource.Resource {
	return &MachinePoolResource{}
}

func (r *MachinePoolResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_machine_pool"
}

func (r *MachinePoolResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "Machine pool.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Must consist of lower-case alphanumeric characters or '-', start and end with an alphanumeric character.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `m5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"replicas": schema.Int64Attribute{
				Description: "The number of machines of the pool",
				Optional:    true,
	***REMOVED***,
			"use_spot_instances": schema.BoolAttribute{
				Description: "Use Amazon EC2 Spot Instances.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"max_spot_price": schema.Float64Attribute{
				Description: "Max Spot price.",
				Optional:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
				Validators: []validator.Float64{
					float64validator.AtLeast(1e-6***REMOVED***, // Greater than zero
		***REMOVED***,
	***REMOVED***,
			"autoscaling_enabled": schema.BoolAttribute{
				Description: "Enables autoscaling. If `true`, this variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
				Optional:    true,
	***REMOVED***,
			"min_replicas": schema.Int64Attribute{
				Description: "The minimum number of replicas for autoscaling functionality.",
				Optional:    true,
	***REMOVED***,
			"max_replicas": schema.Int64Attribute{
				Description: "The maximum number of replicas for autoscaling functionality.",
				Optional:    true,
	***REMOVED***,
			"taints": schema.ListNestedAttribute{
				Description: "Taints for a machine pool. Format should be a comma-separated " +
					"list of 'key=value'. This list will overwrite any modifications " +
					"made to node taints on an ongoing basis.\n",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Taints key",
							Required:    true,
				***REMOVED***,
						"value": schema.StringAttribute{
							Description: "Taints value",
							Required:    true,
				***REMOVED***,
						"schedule_type": schema.StringAttribute{
							Description: "Taints schedule type",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("NoSchedule", "PreferNoSchedule", "NoExecute"***REMOVED***,
					***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				Optional: true,
	***REMOVED***,
			"labels": schema.MapAttribute{
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				ElementType: types.StringType,
				Optional:    true,
	***REMOVED***,
			"multi_availability_zone": schema.BoolAttribute{
				Description: "Create a multi-AZ machine pool for a multi-AZ cluster (default is `true`***REMOVED***",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(***REMOVED***,
					boolplanmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"availability_zone": schema.StringAttribute{
				Description: "Select the availability zone in which to create a single AZ machine pool for a multi-AZ cluster",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"subnet_id": schema.StringAttribute{
				Description: "Select the subnet in which to create a single AZ machine pool for BYO-VPC cluster",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
					stringplanmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"disk_size": schema.Int64Attribute{
				Description: "Root disk size, in GiB.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(***REMOVED***,
					int64planmodifier.UseStateForUnknown(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
	}
}

func (r *MachinePoolResource***REMOVED*** ConfigValidators(context.Context***REMOVED*** []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(path.MatchRoot("availability_zone"***REMOVED***, path.MatchRoot("subnet_id"***REMOVED******REMOVED***,
		resourcevalidator.RequiredTogether(path.MatchRoot("min_replicas"***REMOVED***, path.MatchRoot("max_replicas"***REMOVED******REMOVED***,
		resourcevalidator.Conflicting(path.MatchRoot("replicas"***REMOVED***, path.MatchRoot("min_replicas"***REMOVED******REMOVED***,
		resourcevalidator.Conflicting(path.MatchRoot("replicas"***REMOVED***, path.MatchRoot("max_replicas"***REMOVED******REMOVED***,
	}
}

func (r *MachinePoolResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection***REMOVED***
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData***REMOVED***,
		***REMOVED***
		return
	}

	r.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	state := &MachinePoolState{}
	diags := req.Plan.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	machinepoolName := state.Name.ValueString(***REMOVED***
	if !machinepoolNameRE.MatchString(machinepoolName***REMOVED*** {
		resp.Diagnostics.AddError(
			"Cannot create machine pool: ",
			fmt.Sprintf("Cannot create machine pool for cluster '%s' with name '%s'. Expected a valid value for 'name' matching %s",
				state.Cluster.ValueString(***REMOVED***, state.Name.ValueString(***REMOVED***, machinepoolNameRE,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Wait till the cluster is ready:
	err := common.WaitTillClusterReady(ctx, r.collection, state.Cluster.ValueString(***REMOVED******REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// The default machine pool is created automatically when the cluster is created.
	// We want to import it instead of creating it.
	if machinepoolName == defaultMachinePoolName {
		r.magicImport(ctx, state, resp***REMOVED***
		return
	}

	// Create the machine pool:
	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***
	builder := cmv1.NewMachinePool(***REMOVED***.ID(state.ID.ValueString(***REMOVED******REMOVED***.InstanceType(state.MachineType.ValueString(***REMOVED******REMOVED***
	builder.ID(state.Name.ValueString(***REMOVED******REMOVED***

	if workerDiskSize := common.OptionalInt64(state.DiskSize***REMOVED***; workerDiskSize != nil {
		builder.RootVolume(
			cmv1.NewRootVolume(***REMOVED***.AWS(
				cmv1.NewAWSVolume(***REMOVED***.Size(int(*workerDiskSize***REMOVED******REMOVED***,
			***REMOVED***,
		***REMOVED***
	}

	if err := setSpotInstances(state, builder***REMOVED***; err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s: %v'", state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	isMultiAZPool, err := r.validateAZConfig(state***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	if !common.IsStringAttributeEmpty(state.AvailabilityZone***REMOVED*** {
		builder.AvailabilityZones(state.AvailabilityZone.ValueString(***REMOVED******REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.SubnetID***REMOVED*** {
		builder.Subnets(state.SubnetID.ValueString(***REMOVED******REMOVED***
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg := getAutoscaling(state, builder***REMOVED***
	if errMsg != "" {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s, %s'", state.Cluster.ValueString(***REMOVED***, errMsg,
			***REMOVED***,
		***REMOVED***
		return
	}

	if common.HasValue(state.Replicas***REMOVED*** {
		computeNodeEnabled = true
		if isMultiAZPool && state.Replicas.ValueInt64(***REMOVED***%3 != 0 {
			resp.Diagnostics.AddError(
				"Cannot build machine pool",
				fmt.Sprintf(
					"Cannot build machine pool for cluster '%s', replicas must be a multiple of 3",
					state.Cluster.ValueString(***REMOVED***,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		builder.Replicas(int(state.Replicas.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}
	if (!autoscalingEnabled && !computeNodeEnabled***REMOVED*** || (autoscalingEnabled && computeNodeEnabled***REMOVED*** {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s', please provide a value for either the 'replicas' or 'autoscaling_enabled' parameter. It is mandatory to include at least one of these parameters in the resource plan.",
				state.Cluster.ValueString(***REMOVED***,
			***REMOVED***,
		***REMOVED***
		return
	}

	if state.Taints != nil && len(state.Taints***REMOVED*** > 0 {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range state.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint(***REMOVED***.
				Key(taint.Key.ValueString(***REMOVED******REMOVED***.
				Value(taint.Value.ValueString(***REMOVED******REMOVED***.
				Effect(taint.ScheduleType.ValueString(***REMOVED******REMOVED******REMOVED***
***REMOVED***
		builder.Taints(taintBuilders...***REMOVED***
	}

	if common.HasValue(state.Labels***REMOVED*** {
		labels := map[string]string{}
		for k, v := range state.Labels.Elements(***REMOVED*** {
			labels[k] = v.(types.String***REMOVED***.ValueString(***REMOVED***
***REMOVED***
		builder.Labels(labels***REMOVED***
	}

	object, err := builder.Build(***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	collection := resource.MachinePools(***REMOVED***
	add, err := collection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create machine pool",
			fmt.Sprintf(
				"Cannot create machine pool for cluster '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Save the state:
	r.populateState(object, state***REMOVED***
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

// This handles the "magic" import of the default machine pool, allowing the
// user to include it in their config w/o having to specifically `terraform
// import` it.
func (r *MachinePoolResource***REMOVED*** magicImport(ctx context.Context, state *MachinePoolState, resp *resource.CreateResponse***REMOVED*** {
	machinepoolName := state.Name.ValueString(***REMOVED***
	existingState := &MachinePoolState{
		ID:      types.StringValue(machinepoolName***REMOVED***,
		Cluster: state.Cluster,
		Name:    types.StringValue(machinepoolName***REMOVED***,
	}
	state.ID = types.StringValue(machinepoolName***REMOVED***

	notFound, diags := r.readState(ctx, existingState***REMOVED***
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
				state.Cluster.ValueString(***REMOVED***,
				machinepoolName,
			***REMOVED***,
		***REMOVED***
	}
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}
	diags = r.doUpdate(ctx, existingState, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// the 1AZ-related settings don't apply to the default machine pool
	if state.MultiAvailabilityZone.IsUnknown(***REMOVED*** {
		state.MultiAvailabilityZone = types.BoolNull(***REMOVED***
	}
	if state.AvailabilityZone.IsUnknown(***REMOVED*** {
		state.AvailabilityZone = types.StringNull(***REMOVED***
	}
	if state.SubnetID.IsUnknown(***REMOVED*** {
		state.SubnetID = types.StringNull(***REMOVED***
	}

	// Save the state:
	resp.Diagnostics.Append(resp.State.Set(ctx, state***REMOVED***...***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse***REMOVED*** {
	// Get the current state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	notFound, diags := r.readState(ctx, state***REMOVED***
	if notFound {
		// If we can't find the machine pool, it was deleted. Remove if from the
		// state and don't return an error so the TF apply(***REMOVED*** will automatically
		// recreate it.
		tflog.Warn(ctx, fmt.Sprintf("machine pool (%s***REMOVED*** of cluster (%s***REMOVED*** not found, removing from state",
			state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***,
		***REMOVED******REMOVED***
		resp.State.RemoveResource(ctx***REMOVED***
		return
	}
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state***REMOVED***...***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** readState(ctx context.Context, state *MachinePoolState***REMOVED*** (poolNotFound bool, diags diag.Diagnostics***REMOVED*** {
	diags = diag.Diagnostics{}

	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.ValueString(***REMOVED******REMOVED***
	get, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil && get.Status(***REMOVED*** == http.StatusNotFound {
		poolNotFound = true
		return
	} else if err != nil {
		diags.AddError(
			"Failed to fetch machine pool",
			fmt.Sprintf(
				"Failed to fetch machine pool with identifier %s for cluster %s. Response code: %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, get.Status(***REMOVED***,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := get.Body(***REMOVED***
	r.populateState(object, state***REMOVED***
	return
}

func (r *MachinePoolResource***REMOVED*** Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse***REMOVED*** {
	// Get the state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Get the plan:
	plan := &MachinePoolState{}
	diags = req.Plan.Get(ctx, plan***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	diags = r.doUpdate(ctx, state, plan***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** doUpdate(ctx context.Context, state *MachinePoolState, plan *MachinePoolState***REMOVED*** (diags diag.Diagnostics***REMOVED*** {
	diags = diag.Diagnostics{}

	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.ValueString(***REMOVED******REMOVED***
	_, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***

	if err != nil {
		diags.AddError(
			"Cannot find machine pool",
			fmt.Sprintf(
				"Cannot find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	mpBuilder := cmv1.NewMachinePool(***REMOVED***.ID(state.ID.ValueString(***REMOVED******REMOVED***

	_, ok := common.ShouldPatchString(state.MachineType, plan.MachineType***REMOVED***
	if ok {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s', machine type cannot be updated",
				state.Cluster.ValueString(***REMOVED***,
			***REMOVED***,
		***REMOVED***
		return
	}

	computeNodesEnabled := false
	autoscalingEnabled := false

	if common.HasValue(plan.Replicas***REMOVED*** {
		computeNodesEnabled = true
		mpBuilder.Replicas(int(plan.Replicas.ValueInt64(***REMOVED******REMOVED******REMOVED***

	}

	autoscalingEnabled, errMsg := getAutoscaling(plan, mpBuilder***REMOVED***
	if errMsg != "" {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s, %s ", state.Cluster.ValueString(***REMOVED***, errMsg,
			***REMOVED***,
		***REMOVED***
		return
	}

	if (autoscalingEnabled && computeNodesEnabled***REMOVED*** || (!autoscalingEnabled && !computeNodesEnabled***REMOVED*** {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: either autoscaling or compute nodes should be enabled", state.Cluster.ValueString(***REMOVED***,
			***REMOVED***,
		***REMOVED***
		return
	}

	patchLabels, shouldPatchLabels := common.ShouldPatchMap(state.Labels, plan.Labels***REMOVED***
	if shouldPatchLabels {
		labels := map[string]string{}
		for k, v := range patchLabels.Elements(***REMOVED*** {
			labels[k] = v.(types.String***REMOVED***.ValueString(***REMOVED***
***REMOVED***
		mpBuilder.Labels(labels***REMOVED***
	}

	if shouldPatchTaints(state.Taints, plan.Taints***REMOVED*** {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range plan.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint(***REMOVED***.
				Key(taint.Key.ValueString(***REMOVED******REMOVED***.
				Value(taint.Value.ValueString(***REMOVED******REMOVED***.
				Effect(taint.ScheduleType.ValueString(***REMOVED******REMOVED******REMOVED***
***REMOVED***
		mpBuilder.Taints(taintBuilders...***REMOVED***
	}

	machinePool, err := mpBuilder.Build(***REMOVED***
	if err != nil {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: %v ", state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	update, err := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.ValueString(***REMOVED******REMOVED***.Update(***REMOVED***.Body(machinePool***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		diags.AddError(
			"Failed to update machine pool",
			fmt.Sprintf(
				"Failed to update machine pool '%s'  on cluster '%s': %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := update.Body(***REMOVED***

	// update the autoscaling enabled with the plan value (important for nil and false cases***REMOVED***
	state.AutoScalingEnabled = plan.AutoScalingEnabled
	// update the Replicas with the plan value (important for nil and zero value cases***REMOVED***
	state.Replicas = plan.Replicas

	// Save the state:
	r.populateState(object, state***REMOVED***
	return
}

// Validate the machine pool's settings that pertain to availability zones.
// Returns whether the machine pool is/will be multi-AZ.
func (r *MachinePoolResource***REMOVED*** validateAZConfig(state *MachinePoolState***REMOVED*** (bool, error***REMOVED*** {
	resp, err := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.Get(***REMOVED***.Send(***REMOVED***
	if err != nil {
		return false, fmt.Errorf("failed to get information for cluster %s: %v", state.Cluster.ValueString(***REMOVED***, err***REMOVED***
	}
	cluster := resp.Body(***REMOVED***
	isMultiAZCluster := cluster.MultiAZ(***REMOVED***
	clusterAZs := cluster.Nodes(***REMOVED***.AvailabilityZones(***REMOVED***
	clusterSubnets := cluster.AWS(***REMOVED***.SubnetIDs(***REMOVED***

	if isMultiAZCluster {
		// Can't set both availability_zone and subnet_id
		if !common.IsStringAttributeEmpty(state.AvailabilityZone***REMOVED*** && !common.IsStringAttributeEmpty(state.SubnetID***REMOVED*** {
			return false, fmt.Errorf("availability_zone and subnet_id are mutually exclusive"***REMOVED***
***REMOVED***

		// multi_availability_zone setting must be consistent with availability_zone and subnet_id
		azOrSubnet := !common.IsStringAttributeEmpty(state.AvailabilityZone***REMOVED*** || !common.IsStringAttributeEmpty(state.SubnetID***REMOVED***
		if common.HasValue(state.MultiAvailabilityZone***REMOVED*** {
			if azOrSubnet && state.MultiAvailabilityZone.ValueBool(***REMOVED*** {
				return false, fmt.Errorf("multi_availability_zone must be False when availability_zone or subnet_id is set"***REMOVED***
	***REMOVED***
***REMOVED*** else {
			state.MultiAvailabilityZone = types.BoolValue(!azOrSubnet***REMOVED***
***REMOVED***
	} else { // not a multi-AZ cluster
		if !common.IsStringAttributeEmpty(state.AvailabilityZone***REMOVED*** {
			return false, fmt.Errorf("availability_zone can only be set for multi-AZ clusters"***REMOVED***
***REMOVED***
		if common.HasValue(state.MultiAvailabilityZone***REMOVED*** && state.MultiAvailabilityZone.ValueBool(***REMOVED*** {
			return false, fmt.Errorf("multi_availability_zone can only be set for multi-AZ clusters"***REMOVED***
***REMOVED***
		state.MultiAvailabilityZone = types.BoolValue(false***REMOVED***
	}

	// Ensure that the machine pool's AZ and subnet are valid for the cluster
	// If subnet is set, we make sure it's valid for the cluster, but we don't default it if not set
	if !common.IsStringAttributeEmpty(state.SubnetID***REMOVED*** {
		inClusterSubnet := false
		for _, subnet := range clusterSubnets {
			if subnet == state.SubnetID.ValueString(***REMOVED*** {
				inClusterSubnet = true
				break
	***REMOVED***
***REMOVED***
		if !inClusterSubnet {
			return false, fmt.Errorf("subnet_id %s is not valid for cluster %s", state.SubnetID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED******REMOVED***
***REMOVED***
	} else {
		state.SubnetID = types.StringNull(***REMOVED***
	}
	// If AZ is set, we make sure it's valid for the cluster. If not set and neither is subnet, we default it to the 1st AZ in the cluster
	if !common.IsStringAttributeEmpty(state.AvailabilityZone***REMOVED*** {
		inClusterAZ := false
		for _, az := range clusterAZs {
			if az == state.AvailabilityZone.ValueString(***REMOVED*** {
				inClusterAZ = true
				break
	***REMOVED***
***REMOVED***
		if !inClusterAZ {
			return false, fmt.Errorf("availability_zone %s is not valid for cluster %s", state.AvailabilityZone.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED******REMOVED***
***REMOVED***
	} else {
		if len(clusterAZs***REMOVED*** > 0 && !state.MultiAvailabilityZone.ValueBool(***REMOVED*** && isMultiAZCluster && common.IsStringAttributeEmpty(state.SubnetID***REMOVED*** {
			state.AvailabilityZone = types.StringValue(clusterAZs[0]***REMOVED***
***REMOVED*** else {
			state.AvailabilityZone = types.StringNull(***REMOVED***
***REMOVED***
	}
	return state.MultiAvailabilityZone.ValueBool(***REMOVED***, nil
}

func setSpotInstances(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder***REMOVED*** error {
	useSpotInstances := common.HasValue(state.UseSpotInstances***REMOVED*** && state.UseSpotInstances.ValueBool(***REMOVED***
	isSpotMaxPriceSet := common.HasValue(state.MaxSpotPrice***REMOVED***

	if isSpotMaxPriceSet && !useSpotInstances {
		return errors.New("Cannot set max price when not using spot instances (set \"use_spot_instances\" to true***REMOVED***"***REMOVED***
	}

	if useSpotInstances {
		awsMachinePool := cmv1.NewAWSMachinePool(***REMOVED***
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions(***REMOVED***
		if isSpotMaxPriceSet {
			spotMarketOptions.MaxPrice(state.MaxSpotPrice.ValueFloat64(***REMOVED******REMOVED***
***REMOVED***
		awsMachinePool.SpotMarketOptions(spotMarketOptions***REMOVED***
		mpBuilder.AWS(awsMachinePool***REMOVED***
	}

	return nil
}

func getAutoscaling(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder***REMOVED*** (
	autoscalingEnabled bool, errMsg string***REMOVED*** {
	autoscalingEnabled = false
	if common.HasValue(state.AutoScalingEnabled***REMOVED*** && state.AutoScalingEnabled.ValueBool(***REMOVED*** {
		autoscalingEnabled = true

		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***
		if common.HasValue(state.MaxReplicas***REMOVED*** {
			autoscaling.MaxReplicas(int(state.MaxReplicas.ValueInt64(***REMOVED******REMOVED******REMOVED***
***REMOVED*** else {
			return false, "when enabling autoscaling, should set value for maxReplicas"
***REMOVED***
		if common.HasValue(state.MinReplicas***REMOVED*** {
			autoscaling.MinReplicas(int(state.MinReplicas.ValueInt64(***REMOVED******REMOVED******REMOVED***
***REMOVED*** else {
			return false, "when enabling autoscaling, should set value for minReplicas"
***REMOVED***
		if !autoscaling.Empty(***REMOVED*** {
			mpBuilder.Autoscaling(autoscaling***REMOVED***
***REMOVED***
	} else {
		if common.HasValue(state.MaxReplicas***REMOVED*** || common.HasValue(state.MinReplicas***REMOVED*** {
			return false, "when disabling autoscaling, cannot set min_replicas and/or max_replicas"
***REMOVED***
	}

	return autoscalingEnabled, ""
}

func (r *MachinePoolResource***REMOVED*** Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse***REMOVED*** {
	// Get the state:
	state := &MachinePoolState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the machine pool:
	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.ValueString(***REMOVED******REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		// We can't delete the pool, see if it's the last one:
		numPools, err2 := r.countPools(ctx, state.Cluster.ValueString(***REMOVED******REMOVED***
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
					state.Cluster.ValueString(***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// No return, we want to remove the state
***REMOVED*** else {
			// Wasn't the last one, return error
			resp.Diagnostics.AddError(
				"Cannot delete machine pool",
				fmt.Sprintf(
					"Cannot delete machine pool with identifier '%s' for "+
						"cluster '%s': %v",
					state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}

	// Remove the state:
	resp.State.RemoveResource(ctx***REMOVED***
}

// countPools returns the number of machine pools in the given cluster
func (r *MachinePoolResource***REMOVED*** countPools(ctx context.Context, clusterID string***REMOVED*** (int, error***REMOVED*** {
	resource := r.collection.Cluster(clusterID***REMOVED***.MachinePools(***REMOVED***
	resp, err := resource.List(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		return 0, err
	}
	return resp.Size(***REMOVED***, nil
}

func (r *MachinePoolResource***REMOVED*** ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse***REMOVED*** {
	// To import a machine pool, we need to know the cluster ID and the machine pool ID
	fields := strings.Split(req.ID, ","***REMOVED***
	if len(fields***REMOVED*** != 2 || fields[0] == "" || fields[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Machine pool to import should be specified as <cluster_id>,<machine_pool_id>",
		***REMOVED***
		return
	}
	clusterID := fields[0]
	machinePoolID := fields[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"***REMOVED***, clusterID***REMOVED***...***REMOVED***
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"***REMOVED***, machinePoolID***REMOVED***...***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (r *MachinePoolResource***REMOVED*** populateState(object *cmv1.MachinePool, state *MachinePoolState***REMOVED*** {
	state.ID = types.StringValue(object.ID(***REMOVED******REMOVED***
	state.Name = types.StringValue(object.ID(***REMOVED******REMOVED***

	if getAWS, ok := object.GetAWS(***REMOVED***; ok {
		if spotMarketOptions, ok := getAWS.GetSpotMarketOptions(***REMOVED***; ok {
			state.UseSpotInstances = types.BoolValue(true***REMOVED***
			if spotMarketOptions.MaxPrice(***REMOVED*** != 0 {
				state.MaxSpotPrice = types.Float64Value(spotMarketOptions.MaxPrice(***REMOVED******REMOVED***
	***REMOVED***
***REMOVED***
	}

	autoscaling, ok := object.GetAutoscaling(***REMOVED***
	if ok {
		var minReplicas, maxReplicas int
		state.AutoScalingEnabled = types.BoolValue(true***REMOVED***
		minReplicas, ok = autoscaling.GetMinReplicas(***REMOVED***
		if ok {
			state.MinReplicas = types.Int64Value(int64(minReplicas***REMOVED******REMOVED***
***REMOVED***
		maxReplicas, ok = autoscaling.GetMaxReplicas(***REMOVED***
		if ok {
			state.MaxReplicas = types.Int64Value(int64(maxReplicas***REMOVED******REMOVED***
***REMOVED***
	} else {
		state.MaxReplicas = types.Int64Null(***REMOVED***
		state.MinReplicas = types.Int64Null(***REMOVED***
	}

	if instanceType, ok := object.GetInstanceType(***REMOVED***; ok {
		state.MachineType = types.StringValue(instanceType***REMOVED***
	}

	if replicas, ok := object.GetReplicas(***REMOVED***; ok {
		state.Replicas = types.Int64Value(int64(replicas***REMOVED******REMOVED***
	}

	taints := object.Taints(***REMOVED***
	if len(taints***REMOVED*** > 0 {
		state.Taints = make([]Taints, len(taints***REMOVED******REMOVED***
		for i, taint := range taints {
			state.Taints[i] = Taints{
				Key:          types.StringValue(taint.Key(***REMOVED******REMOVED***,
				Value:        types.StringValue(taint.Value(***REMOVED******REMOVED***,
				ScheduleType: types.StringValue(taint.Effect(***REMOVED******REMOVED***,
	***REMOVED***
***REMOVED***
	} else {
		state.Taints = nil
	}

	labels := object.Labels(***REMOVED***
	if len(labels***REMOVED*** > 0 {
		// XXX: We should be checking error here, but we don't have a way to return the error
		state.Labels, _ = common.ConvertStringMapToMapType(labels***REMOVED***
	} else {
		state.Labels = types.MapNull(types.StringType***REMOVED***
	}

	// Due to RequiresReplace(***REMOVED***, we need to ensure these fields always have a
	// value, even if it's empty. It will be empty if the cluster is multi-AZ or
	// if the other (AZ/subnet***REMOVED*** value is set. We don't need to set
	// MultiAvailibilityZone here, because it's set in the validation function
	// during create.
	state.MultiAvailabilityZone = types.BoolValue(true***REMOVED***
	azs := object.AvailabilityZones(***REMOVED***
	if len(azs***REMOVED*** == 1 {
		state.AvailabilityZone = types.StringValue(azs[0]***REMOVED***
		state.MultiAvailabilityZone = types.BoolValue(false***REMOVED***
	} else {
		state.AvailabilityZone = types.StringValue(""***REMOVED***
	}

	subnets := object.Subnets(***REMOVED***
	if len(subnets***REMOVED*** == 1 {
		state.SubnetID = types.StringValue(subnets[0]***REMOVED***
		state.MultiAvailabilityZone = types.BoolValue(false***REMOVED***
	} else {
		state.SubnetID = types.StringValue(""***REMOVED***
	}

	state.DiskSize = types.Int64Null(***REMOVED***
	if rv, ok := object.GetRootVolume(***REMOVED***; ok {
		if aws, ok := rv.GetAWS(***REMOVED***; ok {
			if workerDiskSize, ok := aws.GetSize(***REMOVED***; ok {
				state.DiskSize = types.Int64Value(int64(workerDiskSize***REMOVED******REMOVED***
	***REMOVED***
***REMOVED***
	}
}

func shouldPatchTaints(a, b []Taints***REMOVED*** bool {
	if (a == nil && b != nil***REMOVED*** || (a != nil && b == nil***REMOVED*** {
		return true
	}
	if len(a***REMOVED*** != len(b***REMOVED*** {
		return true
	}
	for i := range a {
		if !a[i].Key.Equal(b[i].Key***REMOVED*** || !a[i].Value.Equal(b[i].Value***REMOVED*** || !a[i].ScheduleType.Equal(b[i].ScheduleType***REMOVED*** {
			return true
***REMOVED***
	}
	return false
}
