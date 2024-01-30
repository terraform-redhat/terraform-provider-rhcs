/*
Copyright (c) 2023 Red Hat, Inc.

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
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/autoscaler"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type ClusterAutoscalerResourceType struct {
}

type ClusterAutoscalerResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

func New() resource.Resource {
	return &ClusterAutoscalerResource{}
}

var _ resource.Resource = &ClusterAutoscalerResource{}
var _ resource.ResourceWithImportState = &ClusterAutoscalerResource{}
var _ resource.ResourceWithConfigure = &ClusterAutoscalerResource{}

func (r *ClusterAutoscalerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_autoscaler"
}

func (r *ClusterAutoscalerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Cluster-wide autoscaling configuration.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster." + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"balance_similar_node_groups": schema.BoolAttribute{
				Description: "Automatically identify node groups with " +
					"the same instance type and the same set of labels and try " +
					"to keep the respective sizes of those node groups balanced.",
				Optional: true,
			},
			"skip_nodes_with_local_storage": schema.BoolAttribute{
				Description: "If true cluster autoscaler will never delete " +
					"nodes with pods with local storage, e.g. EmptyDir or HostPath. " +
					"true by default at autoscaler.",
				Optional: true,
			},
			"log_verbosity": schema.Int64Attribute{
				Description: "Sets the autoscaler log level. " +
					"Default value is 1, level 4 is recommended for DEBUGGING and " +
					"level 6 will enable almost everything.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"max_pod_grace_period": schema.Int64Attribute{
				Description: "Gives pods graceful termination time before scaling down.",
				Optional:    true,
			},
			"pod_priority_threshold": schema.Int64Attribute{
				Description: "To allow users to schedule 'best-effort' pods, which shouldn't trigger " +
					"Cluster Autoscaler actions, but only run when there are spare resources available.",
				Optional: true,
			},
			"ignore_daemonsets_utilization": schema.BoolAttribute{
				Description: "Should cluster-autoscaler ignore DaemonSet pods when calculating resource utilization " +
					"for scaling down. false by default",
				Optional: true,
			},
			"max_node_provision_time": schema.StringAttribute{
				Description: "Maximum time cluster-autoscaler waits for node to be provisioned.",
				Optional:    true,
				Validators:  []validator.String{autoscaler.DurationStringValidator("max node provision time validation")},
			},
			"balancing_ignored_labels": schema.ListAttribute{
				Description: "This option specifies labels that cluster autoscaler should ignore when " +
					"considering node group similarity. For example, if you have nodes with " +
					"'topology.ebs.csi.aws.com/zone' label, you can add name of this label here " +
					"to prevent cluster autoscaler from splitting nodes into different node groups " +
					"based on its value.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"resource_limits": schema.SingleNestedAttribute{
				Description: "Constraints of autoscaling resources.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"max_nodes_total": schema.Int64Attribute{
						Description: "Maximum number of nodes in all node groups. Cluster autoscaler will " +
							"not grow the cluster beyond this number.",
						Optional: true,
					},
					"cores": autoscaler.RangeAttribute("Minimum and maximum number of cores in cluster, in the format <min>:<max>. "+
						"Cluster autoscaler will not scale the cluster beyond these numbers.", false, true),
					"memory": autoscaler.RangeAttribute("Minimum and maximum number of gigabytes of memory in cluster, in "+
						"the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond "+
						"these numbers.", false, true),
					"gpus": schema.ListNestedAttribute{
						Description: "Minimum and maximum number of different GPUs in cluster, in the format " +
							"<gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond " +
							"these numbers. Can be passed multiple times.",
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required: true,
								},
								"range": autoscaler.RangeAttribute("limit number of GPU type", true, false),
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
				},
			},
			"scale_down": schema.SingleNestedAttribute{
				Description: "Configuration of scale down operation.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Should cluster-autoscaler scale down the cluster.",
						Optional:    true,
					},
					"unneeded_time": schema.StringAttribute{
						Description: "How long a node should be unneeded before it is eligible for scale down.",
						Optional:    true,
					},
					"utilization_threshold": schema.StringAttribute{
						Description: "Node utilization level, defined as sum of requested resources divided " +
							"by capacity, below which a node can be considered for scale down.",
						Optional: true,
						Validators: []validator.String{
							autoscaler.StringFloatRangeValidator("utilization threshold validation", 0.0, 1.0),
						},
					},
					"delay_after_add": schema.StringAttribute{
						Description: "How long after scale up that scale down evaluation resumes.",
						Optional:    true,
						Validators:  []validator.String{autoscaler.DurationStringValidator("delay after add validation")},
					},
					"delay_after_delete": schema.StringAttribute{
						Description: "How long after node deletion that scale down evaluation resumes.",
						Optional:    true,
						Validators:  []validator.String{autoscaler.DurationStringValidator("delay after delete validation")},
					},
					"delay_after_failure": schema.StringAttribute{
						Description: "How long after scale down failure that scale down evaluation resumes.",
						Optional:    true,
						Validators:  []validator.String{autoscaler.DurationStringValidator("delay after failure validation")},
					},
				},
			},
		},
	}
	return
}
func (r *ClusterAutoscalerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	collection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = collection.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection)
}

func (r *ClusterAutoscalerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	plan := &ClusterAutoscalerState{}
	diags := request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Wait till the cluster is ready:
	err := r.clusterWait.WaitForClusterToBeReady(ctx, plan.Cluster.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	autoscaler, err := r.collection.Cluster(plan.Cluster.ValueString()).Autoscaler().Get().Send()
	if err != nil && autoscaler.Status() != http.StatusNotFound {
		response.Diagnostics.AddError("Can't create autoscaler", fmt.Sprintf("Autoscaler for cluster '%s' might already exists. Error: %s",
			plan.Cluster.ValueString(), err.Error()))
		return
	}

	resource := r.collection.Cluster(plan.Cluster.ValueString())

	object, err := clusterAutoscalerStateToObject(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed building cluster autoscaler",
			fmt.Sprintf(
				"Failed building autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	_, err = resource.Autoscaler().Post().Request(object).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed creating cluster autoscaler",
			fmt.Sprintf(
				"Failed creating autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterAutoscalerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	getResponse, err := r.collection.Cluster(state.Cluster.ValueString()).Autoscaler().Get().SendContext(ctx)
	if err != nil && getResponse.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("autoscaler for cluster (%s) not found, removing from state",
			state.Cluster.ValueString(),
		))
		response.State.RemoveResource(ctx)
		return

	} else if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	populateAutoscalerState(getResponse.Body(), state.Cluster.ValueString(), state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterAutoscalerResource) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	var diags diag.Diagnostics

	// Get the state:
	state := &ClusterAutoscalerState{}
	diags = request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &ClusterAutoscalerState{}
	diags = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// assert cluster attribute wasn't changed:
	common.ValidateStateAndPlanEquals(state.Cluster, plan.Cluster, "cluster", &diags)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := r.collection.Cluster(plan.Cluster.ValueString()).Autoscaler().Get().SendContext(ctx)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	autoscaler, err := clusterAutoscalerStateToObject(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed updating cluster autoscaler",
			fmt.Sprintf(
				"Failed updating autoscaler for cluster '%s: %v ",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	update, err := r.collection.Cluster(plan.Cluster.ValueString()).
		Autoscaler().Update().Body(autoscaler).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed updating cluster autoscaler",
			fmt.Sprintf(
				"Failed updating autoscaler for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	object := update.Body()
	state = &ClusterAutoscalerState{}
	populateAutoscalerState(object, plan.Cluster.ValueString(), state)

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (r *ClusterAutoscalerResource) Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse) {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resource := r.collection.Cluster(state.Cluster.ValueString()).Autoscaler()
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed deleting cluster autoscaler",
			fmt.Sprintf(
				"Failed deleting autoscaler for cluster '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return
	}

	response.State.RemoveResource(ctx)
}

func (r *ClusterAutoscalerResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	tflog.Debug(ctx, "begin importstate()")

	resource.ImportStatePassthroughID(ctx, path.Root("cluster"), request, response)
}

// populateAutoscalerState copies the data from the API object to the Terraform state.
func populateAutoscalerState(object *cmv1.ClusterAutoscaler, clusterId string, state *ClusterAutoscalerState) error {
	state.Cluster = types.StringValue(clusterId)

	if value, exists := object.GetBalanceSimilarNodeGroups(); exists {
		state.BalanceSimilarNodeGroups = types.BoolValue(value)
	} else {
		state.BalanceSimilarNodeGroups = types.BoolNull()
	}

	if value, exists := object.GetSkipNodesWithLocalStorage(); exists {
		state.SkipNodesWithLocalStorage = types.BoolValue(value)
	} else {
		state.SkipNodesWithLocalStorage = types.BoolNull()
	}

	if value, exists := object.GetLogVerbosity(); exists {
		state.LogVerbosity = types.Int64Value(int64(value))
	} else {
		state.LogVerbosity = types.Int64Null()
	}

	if value, exists := object.GetMaxPodGracePeriod(); exists {
		state.MaxPodGracePeriod = types.Int64Value(int64(value))
	} else {
		state.MaxPodGracePeriod = types.Int64Null()
	}

	if value, exists := object.GetPodPriorityThreshold(); exists {
		state.PodPriorityThreshold = types.Int64Value(int64(value))
	} else {
		state.PodPriorityThreshold = types.Int64Null()
	}

	if value, exists := object.GetIgnoreDaemonsetsUtilization(); exists {
		state.IgnoreDaemonsetsUtilization = types.BoolValue(value)
	} else {
		state.IgnoreDaemonsetsUtilization = types.BoolNull()
	}

	state.MaxNodeProvisionTime = common.EmptiableStringToStringType(object.MaxNodeProvisionTime())

	if value, exists := object.GetBalancingIgnoredLabels(); exists {
		list, err := common.StringArrayToList(value)
		if err != nil {
			return err
		}
		state.BalancingIgnoredLabels = list
	} else {
		state.BalancingIgnoredLabels = types.ListNull(types.StringType)
	}

	if object.ResourceLimits() != nil {
		state.ResourceLimits = &AutoscalerResourceLimits{}

		if value, exists := object.ResourceLimits().GetMaxNodesTotal(); exists {
			state.ResourceLimits.MaxNodesTotal = types.Int64Value(int64(value))
		} else {
			state.ResourceLimits.MaxNodesTotal = types.Int64Null()
		}

		cores := object.ResourceLimits().Cores()
		if cores != nil {
			state.ResourceLimits.Cores = &autoscaler.ResourceRange{
				Min: types.Int64Value(int64(cores.Min())),
				Max: types.Int64Value(int64(cores.Max())),
			}
		}

		memory := object.ResourceLimits().Memory()
		if memory != nil {
			state.ResourceLimits.Memory = &autoscaler.ResourceRange{
				Min: types.Int64Value(int64(memory.Min())),
				Max: types.Int64Value(int64(memory.Max())),
			}
		}

		gpus := object.ResourceLimits().GPUS()
		if gpus != nil {
			state.ResourceLimits.GPUS = make([]AutoscalerGPULimit, 0)

			for _, gpu := range gpus {
				state.ResourceLimits.GPUS = append(
					state.ResourceLimits.GPUS,
					AutoscalerGPULimit{
						Type: types.StringValue(gpu.Type()),
						Range: autoscaler.ResourceRange{
							Min: types.Int64Value(int64(gpu.Range().Min())),
							Max: types.Int64Value(int64(gpu.Range().Max())),
						},
					},
				)
			}
		}
	}

	if object.ScaleDown() != nil {
		state.ScaleDown = &AutoscalerScaleDownConfig{}

		if value, exists := object.ScaleDown().GetEnabled(); exists {
			state.ScaleDown.Enabled = types.BoolValue(value)
		} else {
			state.ScaleDown.Enabled = types.BoolNull()
		}

		state.ScaleDown.UnneededTime = common.EmptiableStringToStringType(object.ScaleDown().UnneededTime())
		state.ScaleDown.UtilizationThreshold = common.EmptiableStringToStringType(object.ScaleDown().UtilizationThreshold())
		state.ScaleDown.DelayAfterAdd = common.EmptiableStringToStringType(object.ScaleDown().DelayAfterAdd())
		state.ScaleDown.DelayAfterDelete = common.EmptiableStringToStringType(object.ScaleDown().DelayAfterDelete())
		state.ScaleDown.DelayAfterFailure = common.EmptiableStringToStringType(object.ScaleDown().DelayAfterFailure())
	}
	return nil
}

// clusterAutoscalerStateToObject builds a cluster-autoscaler API object from a given Terraform state.
func clusterAutoscalerStateToObject(state *ClusterAutoscalerState) (*cmv1.ClusterAutoscaler, error) {
	builder := cmv1.NewClusterAutoscaler()

	if !state.BalanceSimilarNodeGroups.IsNull() {
		builder.BalanceSimilarNodeGroups(state.BalanceSimilarNodeGroups.ValueBool())
	}

	if !state.SkipNodesWithLocalStorage.IsNull() {
		builder.SkipNodesWithLocalStorage(state.SkipNodesWithLocalStorage.ValueBool())
	}

	if !state.LogVerbosity.IsNull() {
		builder.LogVerbosity(int(state.LogVerbosity.ValueInt64()))
	}

	if !state.MaxPodGracePeriod.IsNull() {
		builder.MaxPodGracePeriod(int(state.MaxPodGracePeriod.ValueInt64()))
	}

	if !state.PodPriorityThreshold.IsNull() {
		builder.PodPriorityThreshold(int(state.PodPriorityThreshold.ValueInt64()))
	}

	if !state.IgnoreDaemonsetsUtilization.IsNull() {
		builder.IgnoreDaemonsetsUtilization(state.IgnoreDaemonsetsUtilization.ValueBool())
	}

	if !state.MaxNodeProvisionTime.IsNull() {
		builder.MaxNodeProvisionTime(state.MaxNodeProvisionTime.ValueString())
	}

	if !state.BalancingIgnoredLabels.IsNull() {
		builder.BalancingIgnoredLabels(common.OptionalList(state.BalancingIgnoredLabels)...)
	}

	if state.ResourceLimits != nil {
		resourceLimitsBuilder := cmv1.NewAutoscalerResourceLimits()

		if !state.ResourceLimits.MaxNodesTotal.IsNull() {
			resourceLimitsBuilder.MaxNodesTotal(int(state.ResourceLimits.MaxNodesTotal.ValueInt64()))
		}

		if state.ResourceLimits.Cores != nil {
			resourceLimitsBuilder.Cores(
				cmv1.NewResourceRange().
					Min(int(state.ResourceLimits.Cores.Min.ValueInt64())).
					Max(int(state.ResourceLimits.Cores.Max.ValueInt64())),
			)
		}

		if state.ResourceLimits.Memory != nil {
			resourceLimitsBuilder.Memory(
				cmv1.NewResourceRange().
					Min(int(state.ResourceLimits.Memory.Min.ValueInt64())).
					Max(int(state.ResourceLimits.Memory.Max.ValueInt64())),
			)
		}

		gpus := make([]*cmv1.AutoscalerResourceLimitsGPULimitBuilder, 0)
		for _, gpu := range state.ResourceLimits.GPUS {
			gpus = append(
				gpus,
				cmv1.NewAutoscalerResourceLimitsGPULimit().
					Type(gpu.Type.ValueString()).
					Range(cmv1.NewResourceRange().
						Min(int(gpu.Range.Min.ValueInt64())).
						Max(int(gpu.Range.Max.ValueInt64()))),
			)
		}
		resourceLimitsBuilder.GPUS(gpus...)

		builder.ResourceLimits(resourceLimitsBuilder)
	}

	if state.ScaleDown != nil {
		scaleDownBuilder := cmv1.NewAutoscalerScaleDownConfig()

		if !state.ScaleDown.Enabled.IsNull() {
			scaleDownBuilder.Enabled(state.ScaleDown.Enabled.ValueBool())
		}

		if !state.ScaleDown.UnneededTime.IsNull() {
			scaleDownBuilder.UnneededTime(state.ScaleDown.UnneededTime.ValueString())
		}

		if !state.ScaleDown.UtilizationThreshold.IsNull() {
			scaleDownBuilder.UtilizationThreshold(state.ScaleDown.UtilizationThreshold.ValueString())
		}

		if !state.ScaleDown.DelayAfterAdd.IsNull() {
			scaleDownBuilder.DelayAfterAdd(state.ScaleDown.DelayAfterAdd.ValueString())
		}

		if !state.ScaleDown.DelayAfterDelete.IsNull() {
			scaleDownBuilder.DelayAfterDelete(state.ScaleDown.DelayAfterDelete.ValueString())
		}

		if !state.ScaleDown.DelayAfterFailure.IsNull() {
			scaleDownBuilder.DelayAfterFailure(state.ScaleDown.DelayAfterFailure.ValueString())
		}

		builder.ScaleDown(scaleDownBuilder)
	}

	return builder.Build()
}
