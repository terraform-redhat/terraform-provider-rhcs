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

package provider

***REMOVED***
	"context"
***REMOVED***
***REMOVED***
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type ClusterAutoscalerResourceType struct {
}

type ClusterAutoscalerResource struct {
	collection *cmv1.ClustersClient
}

func (t *ClusterAutoscalerResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (
	tfsdk.Schema, diag.Diagnostics***REMOVED*** {

	return tfsdk.Schema{
		Description: "Cluster-wide autoscaling configuration.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"balance_similar_node_groups": {
				Description: "Automatically identify node groups with " +
					"the same instance type and the same set of labels and try " +
					"to keep the respective sizes of those node groups balanced.",
				Type:     types.BoolType,
				Optional: true,
	***REMOVED***,
			"skip_nodes_with_local_storage": {
				Description: "If true cluster autoscaler will never delete " +
					"nodes with pods with local storage, e.g. EmptyDir or HostPath. " +
					"true by default at autoscaler.",
				Type:     types.BoolType,
				Optional: true,
	***REMOVED***,
			"log_verbosity": {
				Description: "Sets the autoscaler log level. " +
					"Default value is 1, level 4 is recommended for DEBUGGING and " +
					"level 6 will enable almost everything.",
				Type:     types.Int64Type,
				Optional: true,
	***REMOVED***,
			"max_pod_grace_period": {
				Description: "Gives pods graceful termination time before scaling down.",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"pod_priority_threshold": {
				Description: "To allow users to schedule 'best-effort' pods, which shouldn't trigger " +
					"Cluster Autoscaler actions, but only run when there are spare resources available.",
				Type:     types.Int64Type,
				Optional: true,
	***REMOVED***,
			"ignore_daemonsets_utilization": {
				Description: "Should cluster-autoscaler ignore DaemonSet pods when calculating resource utilization " +
					"for scaling down. false by default",
				Type:     types.BoolType,
				Optional: true,
	***REMOVED***,
			"max_node_provision_time": {
				Description: "Maximum time cluster-autoscaler waits for node to be provisioned.",
				Type:        types.StringType,
				Optional:    true,
				Validators:  DurationStringValidators("max node provision time validation"***REMOVED***,
	***REMOVED***,
			"balancing_ignored_labels": {
				Description: "This option specifies labels that cluster autoscaler should ignore when " +
					"considering node group similarity. For example, if you have nodes with " +
					"'topology.ebs.csi.aws.com/zone' label, you can add name of this label here " +
					"to prevent cluster autoscaler from splitting nodes into different node groups " +
					"based on its value.",
				Type:     types.ListType{ElemType: types.StringType},
				Optional: true,
	***REMOVED***,
			"resource_limits": {
				Description: "Constraints of autoscaling resources.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"max_nodes_total": {
						Description: "Maximum number of nodes in all node groups. Cluster autoscaler will " +
							"not grow the cluster beyond this number.",
						Type:     types.Int64Type,
						Optional: true,
			***REMOVED***,
					"cores": {
						Description: "Minimum and maximum number of cores in cluster, in the format <min>:<max>. " +
							"Cluster autoscaler will not scale the cluster beyond these numbers.",
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"min": {
								Type:     types.Int64Type,
								Required: true,
					***REMOVED***,
							"max": {
								Type:     types.Int64Type,
								Required: true,
					***REMOVED***,
				***REMOVED******REMOVED***,
						Validators: RangeValidators(***REMOVED***,
			***REMOVED***,
					"memory": {
						Description: "Minimum and maximum number of gigabytes of memory in cluster, in " +
							"the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond " +
							"these numbers.",
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"min": {
								Type:     types.Int64Type,
								Required: true,
					***REMOVED***,
							"max": {
								Type:     types.Int64Type,
								Required: true,
					***REMOVED***,
				***REMOVED******REMOVED***,
						Validators: RangeValidators(***REMOVED***,
			***REMOVED***,
					"gpus": {
						Description: "Minimum and maximum number of different GPUs in cluster, in the format " +
							"<gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond " +
							"these numbers. Can be passed multiple times.",
						Optional: true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"type": {
								Type:     types.StringType,
								Required: true,
					***REMOVED***,
							"range": {
								Required: true,
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"min": {
										Type:     types.Int64Type,
										Required: true,
							***REMOVED***,
									"max": {
										Type:     types.Int64Type,
										Required: true,
							***REMOVED***,
						***REMOVED******REMOVED***,
								Validators: RangeValidators(***REMOVED***,
					***REMOVED***,
				***REMOVED***, tfsdk.ListNestedAttributesOptions{
							MinItems: 1,
				***REMOVED******REMOVED***,
			***REMOVED***,
		***REMOVED******REMOVED***,
	***REMOVED***,
			"scale_down": {
				Description: "Configuration of scale down operation.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"enabled": {
						Description: "Should cluster-autoscaler scale down the cluster.",
						Type:        types.BoolType,
						Optional:    true,
			***REMOVED***,
					"unneeded_time": {
						Description: "How long a node should be unneeded before it is eligible for scale down.",
						Type:        types.StringType,
						Optional:    true,
			***REMOVED***,
					"utilization_threshold": {
						Description: "Node utilization level, defined as sum of requested resources divided " +
							"by capacity, below which a node can be considered for scale down.",
						Type:       types.StringType,
						Optional:   true,
						Validators: FloatRangeValidators("utilization threshold validation", 0.0, 1.0***REMOVED***,
			***REMOVED***,
					"delay_after_add": {
						Description: "How long after scale up that scale down evaluation resumes.",
						Type:        types.StringType,
						Optional:    true,
						Validators:  DurationStringValidators("delay after add validation"***REMOVED***,
			***REMOVED***,
					"delay_after_delete": {
						Description: "How long after node deletion that scale down evaluation resumes.",
						Type:        types.StringType,
						Optional:    true,
						Validators:  DurationStringValidators("delay after delete validation"***REMOVED***,
			***REMOVED***,
					"delay_after_failure": {
						Description: "How long after scale down failure that scale down evaluation resumes.",
						Type:        types.StringType,
						Optional:    true,
						Validators:  DurationStringValidators("delay after failure validation"***REMOVED***,
			***REMOVED***,
		***REMOVED******REMOVED***,
	***REMOVED***,
***REMOVED***,
	}, nil
}

func RangeValidators(***REMOVED*** []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "max must be greater or equal to min",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				resourceRange := &AutoscalerResourceRange{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, resourceRange***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				steps := []string{}
				for _, step := range req.AttributePath.Steps(***REMOVED*** {
					steps = append(steps, fmt.Sprintf("%s", step***REMOVED******REMOVED***
		***REMOVED***
				if resourceRange.Min.Value > resourceRange.Max.Value {
					resp.Diagnostics.AddAttributeError(
						req.AttributePath,
						"Invalid resource range",
						fmt.Sprintf("In '%s' attribute, max value must be greater or equal to min value", strings.Join(steps, "."***REMOVED******REMOVED***,
					***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func (t *ClusterAutoscalerResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (tfsdk.Resource, diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	// use it directly when needed.
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	return &ClusterAutoscalerResource{collection: collection}, nil
}

func (r *ClusterAutoscalerResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	plan := &ClusterAutoscalerState{}
	diags := request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	resource := r.collection.Cluster(plan.Cluster.Value***REMOVED***
	// We expect the cluster to exist. Return an error if it's not the case
	if resp, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***; err != nil && resp.Status(***REMOVED*** == http.StatusNotFound {
		message := fmt.Sprintf("Cluster %s not found, error: %v", plan.Cluster.Value, err***REMOVED***
		tflog.Error(ctx, message***REMOVED***
		response.Diagnostics.AddError(
			"Cannot poll cluster state",
			message,
		***REMOVED***
		return
	}

	if err := r.waitForClusterToGetReady(ctx, plan.Cluster.Value***REMOVED***; err != nil {
		response.Diagnostics.AddError(
			"Failed waiting for cluster to get ready",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				plan.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object, err := clusterAutoscalerStateToObject(plan***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed building cluster autoscaler",
			fmt.Sprintf(
				"Failed building autoscaler for cluster '%s': %v",
				plan.Cluster.Value, err,
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
				plan.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	diags = response.State.Set(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	getResponse, err := r.collection.Cluster(state.Cluster.Value***REMOVED***.Autoscaler(***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil && getResponse.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("autoscaler for cluster (%s***REMOVED*** not found, removing from state",
			state.Cluster.Value,
		***REMOVED******REMOVED***
		response.State.RemoveResource(ctx***REMOVED***
		return

	} else if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	populateAutoscalerState(getResponse.Body(***REMOVED***, state.Cluster.Value, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
	var diags diag.Diagnostics

	// Get the plan:
	plan := &ClusterAutoscalerState{}
	diags = request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	_, err := r.collection.Cluster(plan.Cluster.Value***REMOVED***.Autoscaler(***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***

	if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				plan.Cluster.Value, err,
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
				plan.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	update, err := r.collection.Cluster(plan.Cluster.Value***REMOVED***.
		Autoscaler(***REMOVED***.Update(***REMOVED***.Body(autoscaler***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed updating cluster autoscaler",
			fmt.Sprintf(
				"Failed updating autoscaler for cluster '%s': %v",
				plan.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := update.Body(***REMOVED***
	state := &ClusterAutoscalerState{}
	populateAutoscalerState(object, plan.Cluster.Value, state***REMOVED***

	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}
}

func (r *ClusterAutoscalerResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.Autoscaler(***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed deleting cluster autoscaler",
			fmt.Sprintf(
				"Failed deleting autoscaler for cluster '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	clusterId := request.ID

	getResponse, err := r.collection.Cluster(clusterId***REMOVED***.Autoscaler(***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed importing cluster autoscaler",
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}

	state := &ClusterAutoscalerState{}
	populateAutoscalerState(getResponse.Body(***REMOVED***, clusterId, state***REMOVED***
	diags := response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterAutoscalerResource***REMOVED*** waitForClusterToGetReady(ctx context.Context, clusterId string***REMOVED*** error {
	resource := r.collection.Cluster(clusterId***REMOVED***

	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour***REMOVED***
	defer cancel(***REMOVED***

	_, err := resource.Poll(***REMOVED***.
		Interval(30 * time.Second***REMOVED***.
		Predicate(func(get *cmv1.ClusterGetResponse***REMOVED*** bool {
			return get.Body(***REMOVED***.State(***REMOVED*** == cmv1.ClusterStateReady
***REMOVED******REMOVED***.
		StartContext(pollCtx***REMOVED***

	return err
}

// populateAutoscalerState copies the data from the API object to the Terraform state.
func populateAutoscalerState(object *cmv1.ClusterAutoscaler, clusterId string, state *ClusterAutoscalerState***REMOVED*** {
	state.Cluster = types.String{
		Value: clusterId,
	}

	if value, exists := object.GetBalanceSimilarNodeGroups(***REMOVED***; exists {
		state.BalanceSimilarNodeGroups = types.Bool{Value: value}
	} else {
		state.BalanceSimilarNodeGroups = types.Bool{Null: true}
	}

	if value, exists := object.GetSkipNodesWithLocalStorage(***REMOVED***; exists {
		state.SkipNodesWithLocalStorage = types.Bool{Value: value}
	} else {
		state.SkipNodesWithLocalStorage = types.Bool{Null: true}
	}

	if value, exists := object.GetLogVerbosity(***REMOVED***; exists {
		state.LogVerbosity = types.Int64{Value: int64(value***REMOVED***}
	} else {
		state.LogVerbosity = types.Int64{Null: true}
	}

	if value, exists := object.GetMaxPodGracePeriod(***REMOVED***; exists {
		state.MaxPodGracePeriod = types.Int64{Value: int64(value***REMOVED***}
	} else {
		state.MaxPodGracePeriod = types.Int64{Null: true}
	}

	if value, exists := object.GetPodPriorityThreshold(***REMOVED***; exists {
		state.PodPriorityThreshold = types.Int64{Value: int64(value***REMOVED***}
	} else {
		state.PodPriorityThreshold = types.Int64{Null: true}
	}

	if value, exists := object.GetIgnoreDaemonsetsUtilization(***REMOVED***; exists {
		state.IgnoreDaemonsetsUtilization = types.Bool{Value: value}
	} else {
		state.IgnoreDaemonsetsUtilization = types.Bool{Null: true}
	}

	state.MaxNodeProvisionTime = common.EmptiableStringToStringType(object.MaxNodeProvisionTime(***REMOVED******REMOVED***

	if value, exists := object.GetBalancingIgnoredLabels(***REMOVED***; exists {
		state.BalancingIgnoredLabels = common.StringArrayToList(value***REMOVED***
	} else {
		state.BalancingIgnoredLabels = types.List{Null: true}
	}

	if object.ResourceLimits(***REMOVED*** != nil {
		state.ResourceLimits = &AutoscalerResourceLimits{}

		if value, exists := object.ResourceLimits(***REMOVED***.GetMaxNodesTotal(***REMOVED***; exists {
			state.ResourceLimits.MaxNodesTotal = types.Int64{Value: int64(value***REMOVED***}
***REMOVED*** else {
			state.ResourceLimits.MaxNodesTotal = types.Int64{Null: true}
***REMOVED***

		cores := object.ResourceLimits(***REMOVED***.Cores(***REMOVED***
		if cores != nil {
			state.ResourceLimits.Cores = &AutoscalerResourceRange{
				Min: types.Int64{Value: int64(cores.Min(***REMOVED******REMOVED***},
				Max: types.Int64{Value: int64(cores.Max(***REMOVED******REMOVED***},
	***REMOVED***
***REMOVED***

		memory := object.ResourceLimits(***REMOVED***.Memory(***REMOVED***
		if memory != nil {
			state.ResourceLimits.Memory = &AutoscalerResourceRange{
				Min: types.Int64{Value: int64(memory.Min(***REMOVED******REMOVED***},
				Max: types.Int64{Value: int64(memory.Max(***REMOVED******REMOVED***},
	***REMOVED***
***REMOVED***

		gpus := object.ResourceLimits(***REMOVED***.GPUS(***REMOVED***
		if gpus != nil {
			state.ResourceLimits.GPUS = make([]AutoscalerGPULimit, 0***REMOVED***

			for _, gpu := range gpus {
				state.ResourceLimits.GPUS = append(
					state.ResourceLimits.GPUS,
					AutoscalerGPULimit{
						Type: types.String{Value: gpu.Type(***REMOVED***},
						Range: AutoscalerResourceRange{
							Min: types.Int64{Value: int64(gpu.Range(***REMOVED***.Min(***REMOVED******REMOVED***},
							Max: types.Int64{Value: int64(gpu.Range(***REMOVED***.Max(***REMOVED******REMOVED***},
				***REMOVED***,
			***REMOVED***,
				***REMOVED***
	***REMOVED***
***REMOVED***
	}

	if object.ScaleDown(***REMOVED*** != nil {
		state.ScaleDown = &AutoscalerScaleDownConfig{}

		if value, exists := object.ScaleDown(***REMOVED***.GetEnabled(***REMOVED***; exists {
			state.ScaleDown.Enabled = types.Bool{Value: value}
***REMOVED*** else {
			state.ScaleDown.Enabled = types.Bool{Null: true}
***REMOVED***

		state.ScaleDown.UnneededTime = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.UnneededTime(***REMOVED******REMOVED***
		state.ScaleDown.UtilizationThreshold = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.UtilizationThreshold(***REMOVED******REMOVED***
		state.ScaleDown.DelayAfterAdd = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.DelayAfterAdd(***REMOVED******REMOVED***
		state.ScaleDown.DelayAfterDelete = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.DelayAfterDelete(***REMOVED******REMOVED***
		state.ScaleDown.DelayAfterFailure = common.EmptiableStringToStringType(object.ScaleDown(***REMOVED***.DelayAfterFailure(***REMOVED******REMOVED***
	}
}

// clusterAutoscalerStateToObject builds a cluster-autoscaler API object from a given Terraform state.
func clusterAutoscalerStateToObject(state *ClusterAutoscalerState***REMOVED*** (*cmv1.ClusterAutoscaler, error***REMOVED*** {
	builder := cmv1.NewClusterAutoscaler(***REMOVED***

	if !state.BalanceSimilarNodeGroups.Null {
		builder.BalanceSimilarNodeGroups(state.BalanceSimilarNodeGroups.Value***REMOVED***
	}

	if !state.SkipNodesWithLocalStorage.Null {
		builder.SkipNodesWithLocalStorage(state.SkipNodesWithLocalStorage.Value***REMOVED***
	}

	if !state.LogVerbosity.Null {
		builder.LogVerbosity(int(state.LogVerbosity.Value***REMOVED******REMOVED***
	}

	if !state.MaxPodGracePeriod.Null {
		builder.MaxPodGracePeriod(int(state.MaxPodGracePeriod.Value***REMOVED******REMOVED***
	}

	if !state.PodPriorityThreshold.Null {
		builder.PodPriorityThreshold(int(state.PodPriorityThreshold.Value***REMOVED******REMOVED***
	}

	if !state.IgnoreDaemonsetsUtilization.Null {
		builder.IgnoreDaemonsetsUtilization(state.IgnoreDaemonsetsUtilization.Value***REMOVED***
	}

	if !state.MaxNodeProvisionTime.Null {
		builder.MaxNodeProvisionTime(state.MaxNodeProvisionTime.Value***REMOVED***
	}

	if !state.BalancingIgnoredLabels.Null {
		builder.BalancingIgnoredLabels(common.OptionalList(state.BalancingIgnoredLabels***REMOVED***...***REMOVED***
	}

	if state.ResourceLimits != nil {
		resourceLimitsBuilder := cmv1.NewAutoscalerResourceLimits(***REMOVED***

		if !state.ResourceLimits.MaxNodesTotal.Null {
			resourceLimitsBuilder.MaxNodesTotal(int(state.ResourceLimits.MaxNodesTotal.Value***REMOVED******REMOVED***
***REMOVED***

		if state.ResourceLimits.Cores != nil {
			resourceLimitsBuilder.Cores(
				cmv1.NewResourceRange(***REMOVED***.
					Min(int(state.ResourceLimits.Cores.Min.Value***REMOVED******REMOVED***.
					Max(int(state.ResourceLimits.Cores.Max.Value***REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***

		if state.ResourceLimits.Memory != nil {
			resourceLimitsBuilder.Memory(
				cmv1.NewResourceRange(***REMOVED***.
					Min(int(state.ResourceLimits.Memory.Min.Value***REMOVED******REMOVED***.
					Max(int(state.ResourceLimits.Memory.Max.Value***REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***

		gpus := make([]*cmv1.AutoscalerResourceLimitsGPULimitBuilder, 0***REMOVED***
		for _, gpu := range state.ResourceLimits.GPUS {
			gpus = append(
				gpus,
				cmv1.NewAutoscalerResourceLimitsGPULimit(***REMOVED***.
					Type(gpu.Type.Value***REMOVED***.
					Range(cmv1.NewResourceRange(***REMOVED***.
						Min(int(gpu.Range.Min.Value***REMOVED******REMOVED***.
						Max(int(gpu.Range.Max.Value***REMOVED******REMOVED******REMOVED***,
			***REMOVED***
***REMOVED***
		resourceLimitsBuilder.GPUS(gpus...***REMOVED***

		builder.ResourceLimits(resourceLimitsBuilder***REMOVED***
	}

	if state.ScaleDown != nil {
		scaleDownBuilder := cmv1.NewAutoscalerScaleDownConfig(***REMOVED***

		if !state.ScaleDown.Enabled.Null {
			scaleDownBuilder.Enabled(state.ScaleDown.Enabled.Value***REMOVED***
***REMOVED***

		if !state.ScaleDown.UnneededTime.Null {
			scaleDownBuilder.UnneededTime(state.ScaleDown.UnneededTime.Value***REMOVED***
***REMOVED***

		if !state.ScaleDown.UtilizationThreshold.Null {
			scaleDownBuilder.UtilizationThreshold(state.ScaleDown.UtilizationThreshold.Value***REMOVED***
***REMOVED***

		if !state.ScaleDown.DelayAfterAdd.Null {
			scaleDownBuilder.DelayAfterAdd(state.ScaleDown.DelayAfterAdd.Value***REMOVED***
***REMOVED***

		if !state.ScaleDown.DelayAfterDelete.Null {
			scaleDownBuilder.DelayAfterDelete(state.ScaleDown.DelayAfterDelete.Value***REMOVED***
***REMOVED***

		if !state.ScaleDown.DelayAfterFailure.Null {
			scaleDownBuilder.DelayAfterFailure(state.ScaleDown.DelayAfterFailure.Value***REMOVED***
***REMOVED***

		builder.ScaleDown(scaleDownBuilder***REMOVED***
	}

	return builder.Build(***REMOVED***
}
