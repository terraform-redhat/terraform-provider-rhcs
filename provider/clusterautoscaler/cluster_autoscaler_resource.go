/*
Copyright (c***REMOVED*** 2023 Red Hat, Inc.

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

package clusterautoscaler

***REMOVED***
	"context"
***REMOVED***
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type ClusterAutoscalerResourceType struct {
}

type ClusterAutoscalerResource struct {
	collection *cmv1.ClustersClient
}

func New(***REMOVED*** resource.Resource {
	return &ClusterAutoscalerResource{}
}

var _ resource.Resource = &ClusterAutoscalerResource{}
var _ resource.ResourceWithImportState = &ClusterAutoscalerResource{}
var _ resource.ResourceWithConfigure = &ClusterAutoscalerResource{}

func (r *ClusterAutoscalerResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_cluster_autoscaler"
}

func (r *ClusterAutoscalerResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "Cluster-wide autoscaling configuration.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					common.Immutable(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"balance_similar_node_groups": schema.BoolAttribute{
				Description: "Automatically identify node groups with " +
					"the same instance type and the same set of labels and try " +
					"to keep the respective sizes of those node groups balanced.",
				Optional: true,
	***REMOVED***,
			"skip_nodes_with_local_storage": schema.BoolAttribute{
				Description: "If true cluster autoscaler will never delete " +
					"nodes with pods with local storage, e.g. EmptyDir or HostPath. " +
					"true by default at autoscaler.",
				Optional: true,
	***REMOVED***,
			"log_verbosity": schema.Int64Attribute{
				Description: "Sets the autoscaler log level. " +
					"Default value is 1, level 4 is recommended for DEBUGGING and " +
					"level 6 will enable almost everything.",
				Optional: true,
	***REMOVED***,
			"max_pod_grace_period": schema.Int64Attribute{
				Description: "Gives pods graceful termination time before scaling down.",
				Optional:    true,
	***REMOVED***,
			"pod_priority_threshold": schema.Int64Attribute{
				Description: "To allow users to schedule 'best-effort' pods, which shouldn't trigger " +
					"Cluster Autoscaler actions, but only run when there are spare resources available.",
				Optional: true,
	***REMOVED***,
			"ignore_daemonsets_utilization": schema.BoolAttribute{
				Description: "Should cluster-autoscaler ignore DaemonSet pods when calculating resource utilization " +
					"for scaling down. false by default",
				Optional: true,
	***REMOVED***,
			"max_node_provision_time": schema.StringAttribute{
				Description: "Maximum time cluster-autoscaler waits for node to be provisioned.",
				Optional:    true,
				Validators:  []validator.String{durationStringValidator("max node provision time validation"***REMOVED***},
	***REMOVED***,
			"balancing_ignored_labels": schema.ListAttribute{
				Description: "This option specifies labels that cluster autoscaler should ignore when " +
					"considering node group similarity. For example, if you have nodes with " +
					"'topology.ebs.csi.aws.com/zone' label, you can add name of this label here " +
					"to prevent cluster autoscaler from splitting nodes into different node groups " +
					"based on its value.",
				ElementType: types.StringType,
				Optional:    true,
	***REMOVED***,
			"resource_limits": schema.SingleNestedAttribute{
				Description: "Constraints of autoscaling resources.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"max_nodes_total": schema.Int64Attribute{
						Description: "Maximum number of nodes in all node groups. Cluster autoscaler will " +
							"not grow the cluster beyond this number.",
						Optional: true,
			***REMOVED***,
					"cores": rangeAttribute("Minimum and maximum number of cores in cluster, in the format <min>:<max>. "+
						"Cluster autoscaler will not scale the cluster beyond these numbers.", false, true***REMOVED***,
					"memory": rangeAttribute("Minimum and maximum number of gigabytes of memory in cluster, in "+
						"the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond "+
						"these numbers.", false, true***REMOVED***,
					"gpus": schema.ListNestedAttribute{
						Description: "Minimum and maximum number of different GPUs in cluster, in the format " +
							"<gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond " +
							"these numbers. Can be passed multiple times.",
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required: true,
						***REMOVED***,
								"range": rangeAttribute("limit number of GPU type", true, false***REMOVED***,
					***REMOVED***,
				***REMOVED***,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1***REMOVED***,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"scale_down": schema.SingleNestedAttribute{
				Description: "Configuration of scale down operation.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Should cluster-autoscaler scale down the cluster.",
						Optional:    true,
			***REMOVED***,
					"unneeded_time": schema.StringAttribute{
						Description: "How long a node should be unneeded before it is eligible for scale down.",
						Optional:    true,
			***REMOVED***,
					"utilization_threshold": schema.StringAttribute{
						Description: "Node utilization level, defined as sum of requested resources divided " +
							"by capacity, below which a node can be considered for scale down.",
						Optional: true,
						Validators: []validator.String{
							stringFloatRangeValidator("utilization threshold validation", 0.0, 1.0***REMOVED***,
				***REMOVED***,
			***REMOVED***,
					"delay_after_add": schema.StringAttribute{
						Description: "How long after scale up that scale down evaluation resumes.",
						Optional:    true,
						Validators:  []validator.String{durationStringValidator("delay after add validation"***REMOVED***},
			***REMOVED***,
					"delay_after_delete": schema.StringAttribute{
						Description: "How long after node deletion that scale down evaluation resumes.",
						Optional:    true,
						Validators:  []validator.String{durationStringValidator("delay after delete validation"***REMOVED***},
			***REMOVED***,
					"delay_after_failure": schema.StringAttribute{
						Description: "How long after scale down failure that scale down evaluation resumes.",
						Optional:    true,
						Validators:  []validator.String{durationStringValidator("delay after failure validation"***REMOVED***},
			***REMOVED***,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
	}
	return
}
func (r *ClusterAutoscalerResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	if req.ProviderData == nil {
		return
	}

	collection, ok := req.ProviderData.(*sdk.Connection***REMOVED***
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData***REMOVED***,
		***REMOVED***
		return
	}

	r.collection = collection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse***REMOVED*** {
	plan := &ClusterAutoscalerState{}
	diags := request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Wait till the cluster is ready:
	err := common.WaitTillClusterReady(ctx, r.collection, plan.Cluster.ValueString(***REMOVED******REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				plan.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	autoscaler, err := r.collection.Cluster(plan.Cluster.ValueString(***REMOVED******REMOVED***.Autoscaler(***REMOVED***.Get(***REMOVED***.Send(***REMOVED***
	if err != nil && autoscaler.Status(***REMOVED*** != http.StatusNotFound {
		response.Diagnostics.AddError("Can't create autoscaler", fmt.Sprintf("Autoscaler for cluster '%s' might already exists. Error: %s",
			plan.Cluster.ValueString(***REMOVED***, err.Error(***REMOVED******REMOVED******REMOVED***
		return
	}

	resource := r.collection.Cluster(plan.Cluster.ValueString(***REMOVED******REMOVED***

	object, err := clusterAutoscalerStateToObject(plan***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed building cluster autoscaler",
			fmt.Sprintf(
				"Failed building autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	_, err = resource.Autoscaler(***REMOVED***.Post(***REMOVED***.Request(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed creating cluster autoscaler",
			fmt.Sprintf(
				"Failed creating autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	diags = response.State.Set(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse***REMOVED*** {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	getResponse, err := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.Autoscaler(***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil && getResponse.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("autoscaler for cluster (%s***REMOVED*** not found, removing from state",
			state.Cluster.ValueString(***REMOVED***,
		***REMOVED******REMOVED***
		response.State.RemoveResource(ctx***REMOVED***
		return

	} else if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	populateAutoscalerState(getResponse.Body(***REMOVED***, state.Cluster.ValueString(***REMOVED***, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse***REMOVED*** {
	var diags diag.Diagnostics

	// Get the plan:
	plan := &ClusterAutoscalerState{}
	diags = request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	_, err := r.collection.Cluster(plan.Cluster.ValueString(***REMOVED******REMOVED***.Autoscaler(***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***

	if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	autoscaler, err := clusterAutoscalerStateToObject(plan***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed updating cluster autoscaler",
			fmt.Sprintf(
				"Failed updating autoscaler for cluster '%s: %v ",
				plan.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	update, err := r.collection.Cluster(plan.Cluster.ValueString(***REMOVED******REMOVED***.
		Autoscaler(***REMOVED***.Update(***REMOVED***.Body(autoscaler***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed updating cluster autoscaler",
			fmt.Sprintf(
				"Failed updating autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := update.Body(***REMOVED***
	state := &ClusterAutoscalerState{}
	populateAutoscalerState(object, plan.Cluster.ValueString(***REMOVED***, state***REMOVED***

	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}
}

func (r *ClusterAutoscalerResource***REMOVED*** Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse***REMOVED*** {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.Autoscaler(***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed deleting cluster autoscaler",
			fmt.Sprintf(
				"Failed deleting autoscaler for cluster '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse***REMOVED*** {
	tflog.Debug(ctx, "begin importstate(***REMOVED***"***REMOVED***

	resource.ImportStatePassthroughID(ctx, path.Root("cluster"***REMOVED***, request, response***REMOVED***
}

// populateAutoscalerState copies the data from the API object to the Terraform state.
func populateAutoscalerState(object *cmv1.ClusterAutoscaler, clusterId string, state *ClusterAutoscalerState***REMOVED*** error {
	state.Cluster = types.StringValue(clusterId***REMOVED***

	if value, exists := object.GetBalanceSimilarNodeGroups(***REMOVED***; exists {
		state.BalanceSimilarNodeGroups = types.BoolValue(value***REMOVED***
	} else {
		state.BalanceSimilarNodeGroups = types.BoolNull(***REMOVED***
	}

	if value, exists := object.GetSkipNodesWithLocalStorage(***REMOVED***; exists {
		state.SkipNodesWithLocalStorage = types.BoolValue(value***REMOVED***
	} else {
		state.SkipNodesWithLocalStorage = types.BoolNull(***REMOVED***
	}

	if value, exists := object.GetLogVerbosity(***REMOVED***; exists {
		state.LogVerbosity = types.Int64Value(int64(value***REMOVED******REMOVED***
	} else {
		state.LogVerbosity = types.Int64Null(***REMOVED***
	}

	if value, exists := object.GetMaxPodGracePeriod(***REMOVED***; exists {
		state.MaxPodGracePeriod = types.Int64Value(int64(value***REMOVED******REMOVED***
	} else {
		state.MaxPodGracePeriod = types.Int64Null(***REMOVED***
	}

	if value, exists := object.GetPodPriorityThreshold(***REMOVED***; exists {
		state.PodPriorityThreshold = types.Int64Value(int64(value***REMOVED******REMOVED***
	} else {
		state.PodPriorityThreshold = types.Int64Null(***REMOVED***
	}

	if value, exists := object.GetIgnoreDaemonsetsUtilization(***REMOVED***; exists {
		state.IgnoreDaemonsetsUtilization = types.BoolValue(value***REMOVED***
	} else {
		state.IgnoreDaemonsetsUtilization = types.BoolNull(***REMOVED***
	}

	state.MaxNodeProvisionTime = common.EmptiableStringToStringType(object.MaxNodeProvisionTime(***REMOVED******REMOVED***

	if value, exists := object.GetBalancingIgnoredLabels(***REMOVED***; exists {
		list, err := common.StringArrayToList(value***REMOVED***
		if err != nil {
			return err
***REMOVED***
		state.BalancingIgnoredLabels = list
	} else {
		state.BalancingIgnoredLabels = types.ListNull(types.StringType***REMOVED***
	}

	if object.ResourceLimits(***REMOVED*** != nil {
		state.ResourceLimits = &AutoscalerResourceLimits{}

		if value, exists := object.ResourceLimits(***REMOVED***.GetMaxNodesTotal(***REMOVED***; exists {
			state.ResourceLimits.MaxNodesTotal = types.Int64Value(int64(value***REMOVED******REMOVED***
***REMOVED*** else {
			state.ResourceLimits.MaxNodesTotal = types.Int64Null(***REMOVED***
***REMOVED***

		cores := object.ResourceLimits(***REMOVED***.Cores(***REMOVED***
		if cores != nil {
			state.ResourceLimits.Cores = &AutoscalerResourceRange{
				Min: types.Int64Value(int64(cores.Min(***REMOVED******REMOVED******REMOVED***,
				Max: types.Int64Value(int64(cores.Max(***REMOVED******REMOVED******REMOVED***,
	***REMOVED***
***REMOVED***

		memory := object.ResourceLimits(***REMOVED***.Memory(***REMOVED***
		if memory != nil {
			state.ResourceLimits.Memory = &AutoscalerResourceRange{
				Min: types.Int64Value(int64(memory.Min(***REMOVED******REMOVED******REMOVED***,
				Max: types.Int64Value(int64(memory.Max(***REMOVED******REMOVED******REMOVED***,
	***REMOVED***
***REMOVED***

		gpus := object.ResourceLimits(***REMOVED***.GPUS(***REMOVED***
		if gpus != nil {
			state.ResourceLimits.GPUS = make([]AutoscalerGPULimit, 0***REMOVED***

			for _, gpu := range gpus {
				state.ResourceLimits.GPUS = append(
					state.ResourceLimits.GPUS,
					AutoscalerGPULimit{
						Type: types.StringValue(gpu.Type(***REMOVED******REMOVED***,
						Range: AutoscalerResourceRange{
							Min: types.Int64Value(int64(gpu.Range(***REMOVED***.Min(***REMOVED******REMOVED******REMOVED***,
							Max: types.Int64Value(int64(gpu.Range(***REMOVED***.Max(***REMOVED******REMOVED******REMOVED***,
				***REMOVED***,
			***REMOVED***,
				***REMOVED***
	***REMOVED***
***REMOVED***
	}

	if object.ScaleDown(***REMOVED*** != nil {
		state.ScaleDown = &AutoscalerScaleDownConfig{}

		if value, exists := object.ScaleDown(***REMOVED***.GetEnabled(***REMOVED***; exists {
			state.ScaleDown.Enabled = types.BoolValue(value***REMOVED***
***REMOVED*** else {
			state.ScaleDown.Enabled = types.BoolNull(***REMOVED***
***REMOVED***

		state.ScaleDown.UnneededTime = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.UnneededTime(***REMOVED******REMOVED***
		state.ScaleDown.UtilizationThreshold = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.UtilizationThreshold(***REMOVED******REMOVED***
		state.ScaleDown.DelayAfterAdd = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.DelayAfterAdd(***REMOVED******REMOVED***
		state.ScaleDown.DelayAfterDelete = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.DelayAfterDelete(***REMOVED******REMOVED***
		state.ScaleDown.DelayAfterFailure = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.DelayAfterFailure(***REMOVED******REMOVED***
	}
	return nil
}

// clusterAutoscalerStateToObject builds a cluster-autoscaler API object from a given Terraform state.
func clusterAutoscalerStateToObject(state *ClusterAutoscalerState***REMOVED*** (*cmv1.ClusterAutoscaler, error***REMOVED*** {
	builder := cmv1.NewClusterAutoscaler(***REMOVED***

	if !state.BalanceSimilarNodeGroups.IsNull(***REMOVED*** {
		builder.BalanceSimilarNodeGroups(state.BalanceSimilarNodeGroups.ValueBool(***REMOVED******REMOVED***
	}

	if !state.SkipNodesWithLocalStorage.IsNull(***REMOVED*** {
		builder.SkipNodesWithLocalStorage(state.SkipNodesWithLocalStorage.ValueBool(***REMOVED******REMOVED***
	}

	if !state.LogVerbosity.IsNull(***REMOVED*** {
		builder.LogVerbosity(int(state.LogVerbosity.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}

	if !state.MaxPodGracePeriod.IsNull(***REMOVED*** {
		builder.MaxPodGracePeriod(int(state.MaxPodGracePeriod.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}

	if !state.PodPriorityThreshold.IsNull(***REMOVED*** {
		builder.PodPriorityThreshold(int(state.PodPriorityThreshold.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}

	if !state.IgnoreDaemonsetsUtilization.IsNull(***REMOVED*** {
		builder.IgnoreDaemonsetsUtilization(state.IgnoreDaemonsetsUtilization.ValueBool(***REMOVED******REMOVED***
	}

	if !state.MaxNodeProvisionTime.IsNull(***REMOVED*** {
		builder.MaxNodeProvisionTime(state.MaxNodeProvisionTime.ValueString(***REMOVED******REMOVED***
	}

	if !state.BalancingIgnoredLabels.IsNull(***REMOVED*** {
		builder.BalancingIgnoredLabels(common.OptionalList(state.BalancingIgnoredLabels***REMOVED***...***REMOVED***
	}

	if state.ResourceLimits != nil {
		resourceLimitsBuilder := cmv1.NewAutoscalerResourceLimits(***REMOVED***

		if !state.ResourceLimits.MaxNodesTotal.IsNull(***REMOVED*** {
			resourceLimitsBuilder.MaxNodesTotal(int(state.ResourceLimits.MaxNodesTotal.ValueInt64(***REMOVED******REMOVED******REMOVED***
***REMOVED***

		if state.ResourceLimits.Cores != nil {
			resourceLimitsBuilder.Cores(
				cmv1.NewResourceRange(***REMOVED***.
					Min(int(state.ResourceLimits.Cores.Min.ValueInt64(***REMOVED******REMOVED******REMOVED***.
					Max(int(state.ResourceLimits.Cores.Max.ValueInt64(***REMOVED******REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***

		if state.ResourceLimits.Memory != nil {
			resourceLimitsBuilder.Memory(
				cmv1.NewResourceRange(***REMOVED***.
					Min(int(state.ResourceLimits.Memory.Min.ValueInt64(***REMOVED******REMOVED******REMOVED***.
					Max(int(state.ResourceLimits.Memory.Max.ValueInt64(***REMOVED******REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***

		gpus := make([]*cmv1.AutoscalerResourceLimitsGPULimitBuilder, 0***REMOVED***
		for _, gpu := range state.ResourceLimits.GPUS {
			gpus = append(
				gpus,
				cmv1.NewAutoscalerResourceLimitsGPULimit(***REMOVED***.
					Type(gpu.Type.ValueString(***REMOVED******REMOVED***.
					Range(cmv1.NewResourceRange(***REMOVED***.
						Min(int(gpu.Range.Min.ValueInt64(***REMOVED******REMOVED******REMOVED***.
						Max(int(gpu.Range.Max.ValueInt64(***REMOVED******REMOVED******REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***
		resourceLimitsBuilder.GPUS(gpus...***REMOVED***

		builder.ResourceLimits(resourceLimitsBuilder***REMOVED***
	}

	if state.ScaleDown != nil {
		scaleDownBuilder := cmv1.NewAutoscalerScaleDownConfig(***REMOVED***

		if !state.ScaleDown.Enabled.IsNull(***REMOVED*** {
			scaleDownBuilder.Enabled(state.ScaleDown.Enabled.ValueBool(***REMOVED******REMOVED***
***REMOVED***

		if !state.ScaleDown.UnneededTime.IsNull(***REMOVED*** {
			scaleDownBuilder.UnneededTime(state.ScaleDown.UnneededTime.ValueString(***REMOVED******REMOVED***
***REMOVED***

		if !state.ScaleDown.UtilizationThreshold.IsNull(***REMOVED*** {
			scaleDownBuilder.UtilizationThreshold(state.ScaleDown.UtilizationThreshold.ValueString(***REMOVED******REMOVED***
***REMOVED***

		if !state.ScaleDown.DelayAfterAdd.IsNull(***REMOVED*** {
			scaleDownBuilder.DelayAfterAdd(state.ScaleDown.DelayAfterAdd.ValueString(***REMOVED******REMOVED***
***REMOVED***

		if !state.ScaleDown.DelayAfterDelete.IsNull(***REMOVED*** {
			scaleDownBuilder.DelayAfterDelete(state.ScaleDown.DelayAfterDelete.ValueString(***REMOVED******REMOVED***
***REMOVED***

		if !state.ScaleDown.DelayAfterFailure.IsNull(***REMOVED*** {
			scaleDownBuilder.DelayAfterFailure(state.ScaleDown.DelayAfterFailure.ValueString(***REMOVED******REMOVED***
***REMOVED***

		builder.ScaleDown(scaleDownBuilder***REMOVED***
	}

	return builder.Build(***REMOVED***
}
