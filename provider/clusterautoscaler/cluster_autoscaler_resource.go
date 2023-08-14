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

package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type ClusterAutoscalerResourceType struct {
}

type ClusterAutoscalerResource struct {
	collection *cmv1.ClustersClient
}

func (t *ClusterAutoscalerResourceType) GetSchema(ctx context.Context) (
	tfsdk.Schema, diag.Diagnostics) {

	return tfsdk.Schema{
		Description: "Cluster-wide autoscaling configuration.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"balance_similar_node_groups": {
				Description: "Automatically identify node groups with " +
					"the same instance type and the same set of labels and try " +
					"to keep the respective sizes of those node groups balanced.",
				Type:     types.BoolType,
				Optional: true,
			},
			"skip_nodes_with_local_storage": {
				Description: "If true cluster autoscaler will never delete " +
					"nodes with pods with local storage, e.g. EmptyDir or HostPath. " +
					"true by default at autoscaler.",
				Type:     types.BoolType,
				Optional: true,
			},
			"log_verbosity": {
				Description: "Sets the autoscaler log level. " +
					"Default value is 1, level 4 is recommended for DEBUGGING and " +
					"level 6 will enable almost everything.",
				Type:     types.Int64Type,
				Optional: true,
			},
			"max_pod_grace_period": {
				Description: "Gives pods graceful termination time before scaling down.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"pod_priority_threshold": {
				Description: "To allow users to schedule 'best-effort' pods, which shouldn't trigger " +
					"Cluster Autoscaler actions, but only run when there are spare resources available.",
				Type:     types.Int64Type,
				Optional: true,
			},
			"ignore_daemonsets_utilization": {
				Description: "Should cluster-autoscaler ignore DaemonSet pods when calculating resource utilization " +
					"for scaling down. false by default",
				Type:     types.BoolType,
				Optional: true,
			},
			"max_node_provision_time": {
				Description: "Maximum time cluster-autoscaler waits for node to be provisioned.",
				Type:        types.StringType,
				Optional:    true,
				Validators:  DurationStringValidators("max node provision time validation"),
			},
			"balancing_ignored_labels": {
				Description: "This option specifies labels that cluster autoscaler should ignore when " +
					"considering node group similarity. For example, if you have nodes with " +
					"'topology.ebs.csi.aws.com/zone' label, you can add name of this label here " +
					"to prevent cluster autoscaler from splitting nodes into different node groups " +
					"based on its value.",
				Type:     types.ListType{ElemType: types.StringType},
				Optional: true,
			},
			"resource_limits": {
				Description: "Constraints of autoscaling resources.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"max_nodes_total": {
						Description: "Maximum number of nodes in all node groups. Cluster autoscaler will " +
							"not grow the cluster beyond this number.",
						Type:     types.Int64Type,
						Optional: true,
					},
					"cores": {
						Description: "Minimum and maximum number of cores in cluster, in the format <min>:<max>. " +
							"Cluster autoscaler will not scale the cluster beyond these numbers.",
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"min": {
								Type:     types.Int64Type,
								Required: true,
							},
							"max": {
								Type:     types.Int64Type,
								Required: true,
							},
						}),
						Validators: RangeValidators(),
					},
					"memory": {
						Description: "Minimum and maximum number of gigabytes of memory in cluster, in " +
							"the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond " +
							"these numbers.",
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"min": {
								Type:     types.Int64Type,
								Required: true,
							},
							"max": {
								Type:     types.Int64Type,
								Required: true,
							},
						}),
						Validators: RangeValidators(),
					},
					"gpus": {
						Description: "Minimum and maximum number of different GPUs in cluster, in the format " +
							"<gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond " +
							"these numbers. Can be passed multiple times.",
						Optional: true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"type": {
								Type:     types.StringType,
								Required: true,
							},
							"range": {
								Required: true,
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"min": {
										Type:     types.Int64Type,
										Required: true,
									},
									"max": {
										Type:     types.Int64Type,
										Required: true,
									},
								}),
								Validators: RangeValidators(),
							},
						}, tfsdk.ListNestedAttributesOptions{
							MinItems: 1,
						}),
					},
				}),
			},
			"scale_down": {
				Description: "Configuration of scale down operation.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"enabled": {
						Description: "Should cluster-autoscaler scale down the cluster.",
						Type:        types.BoolType,
						Optional:    true,
					},
					"unneeded_time": {
						Description: "How long a node should be unneeded before it is eligible for scale down.",
						Type:        types.StringType,
						Optional:    true,
					},
					"utilization_threshold": {
						Description: "Node utilization level, defined as sum of requested resources divided " +
							"by capacity, below which a node can be considered for scale down.",
						Type:       types.StringType,
						Optional:   true,
						Validators: FloatRangeValidators("utilization threshold validation", 0.0, 1.0),
					},
					"delay_after_add": {
						Description: "How long after scale up that scale down evaluation resumes.",
						Type:        types.StringType,
						Optional:    true,
						Validators:  DurationStringValidators("delay after add validation"),
					},
					"delay_after_delete": {
						Description: "How long after node deletion that scale down evaluation resumes.",
						Type:        types.StringType,
						Optional:    true,
						Validators:  DurationStringValidators("delay after delete validation"),
					},
					"delay_after_failure": {
						Description: "How long after scale down failure that scale down evaluation resumes.",
						Type:        types.StringType,
						Optional:    true,
						Validators:  DurationStringValidators("delay after failure validation"),
					},
				}),
			},
		},
	}, nil
}

func RangeValidators() []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "max must be greater or equal to min",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				resourceRange := &AutoscalerResourceRange{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, resourceRange)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				steps := []string{}
				for _, step := range req.AttributePath.Steps() {
					steps = append(steps, fmt.Sprintf("%s", step))
				}
				if resourceRange.Min.Value > resourceRange.Max.Value {
					resp.Diagnostics.AddAttributeError(
						req.AttributePath,
						"Invalid resource range",
						fmt.Sprintf("In '%s' attribute, max value must be greater or equal to min value", strings.Join(steps, ".")),
					)
				}
			},
		},
	}
}

func (t *ClusterAutoscalerResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	// Cast the provider interface to the specific implementation:
	// use it directly when needed.
	parent := p.(*Provider)

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt().V1().Clusters()

	// Create the resource:
	return &ClusterAutoscalerResource{collection: collection}, nil
}

func (r *ClusterAutoscalerResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	plan := &ClusterAutoscalerState{}
	diags := request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resource := r.collection.Cluster(plan.Cluster.Value)
	// We expect the cluster to exist. Return an error if it's not the case
	if resp, err := resource.Get().SendContext(ctx); err != nil && resp.Status() == http.StatusNotFound {
		message := fmt.Sprintf("Cluster %s not found, error: %v", plan.Cluster.Value, err)
		tflog.Error(ctx, message)
		response.Diagnostics.AddError(
			"Cannot poll cluster state",
			message,
		)
		return
	}

	if err := r.waitForClusterToGetReady(ctx, plan.Cluster.Value); err != nil {
		response.Diagnostics.AddError(
			"Failed waiting for cluster to get ready",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				plan.Cluster.Value, err,
			),
		)
		return
	}

	object, err := clusterAutoscalerStateToObject(plan)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed building cluster autoscaler",
			fmt.Sprintf(
				"Failed building autoscaler for cluster '%s': %v",
				plan.Cluster.Value, err,
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
				plan.Cluster.Value, err,
			),
		)
		return
	}

	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterAutoscalerResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	getResponse, err := r.collection.Cluster(state.Cluster.Value).Autoscaler().Get().SendContext(ctx)
	if err != nil && getResponse.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("autoscaler for cluster (%s) not found, removing from state",
			state.Cluster.Value,
		))
		response.State.RemoveResource(ctx)
		return

	} else if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	populateAutoscalerState(getResponse.Body(), state.Cluster.Value, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterAutoscalerResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	var diags diag.Diagnostics

	// Get the plan:
	plan := &ClusterAutoscalerState{}
	diags = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := r.collection.Cluster(plan.Cluster.Value).Autoscaler().Get().SendContext(ctx)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed getting cluster autoscaler",
			fmt.Sprintf(
				"Failed getting autoscaler for cluster '%s': %v",
				plan.Cluster.Value, err,
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
				plan.Cluster.Value, err,
			),
		)
		return
	}

	update, err := r.collection.Cluster(plan.Cluster.Value).
		Autoscaler().Update().Body(autoscaler).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed updating cluster autoscaler",
			fmt.Sprintf(
				"Failed updating autoscaler for cluster '%s': %v",
				plan.Cluster.Value, err,
			),
		)
		return
	}

	object := update.Body()
	state := &ClusterAutoscalerState{}
	populateAutoscalerState(object, plan.Cluster.Value, state)

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (r *ClusterAutoscalerResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	state := &ClusterAutoscalerState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resource := r.collection.Cluster(state.Cluster.Value).Autoscaler()
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed deleting cluster autoscaler",
			fmt.Sprintf(
				"Failed deleting autoscaler for cluster '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}

	response.State.RemoveResource(ctx)
}

func (r *ClusterAutoscalerResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	clusterId := request.ID

	getResponse, err := r.collection.Cluster(clusterId).Autoscaler().Get().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed importing cluster autoscaler",
			err.Error(),
		)
		return
	}

	state := &ClusterAutoscalerState{}
	populateAutoscalerState(getResponse.Body(), clusterId, state)
	diags := response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterAutoscalerResource) waitForClusterToGetReady(ctx context.Context, clusterId string) error {
	resource := r.collection.Cluster(clusterId)

	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()

	_, err := resource.Poll().
		Interval(30 * time.Second).
		Predicate(func(get *cmv1.ClusterGetResponse) bool {
			return get.Body().State() == cmv1.ClusterStateReady
		}).
		StartContext(pollCtx)

	return err
}

// populateAutoscalerState copies the data from the API object to the Terraform state.
func populateAutoscalerState(object *cmv1.ClusterAutoscaler, clusterId string, state *ClusterAutoscalerState) {
	state.Cluster = types.String{
		Value: clusterId,
	}

	if value, exists := object.GetBalanceSimilarNodeGroups(); exists {
		state.BalanceSimilarNodeGroups = types.Bool{Value: value}
	} else {
		state.BalanceSimilarNodeGroups = types.Bool{Null: true}
	}

	if value, exists := object.GetSkipNodesWithLocalStorage(); exists {
		state.SkipNodesWithLocalStorage = types.Bool{Value: value}
	} else {
		state.SkipNodesWithLocalStorage = types.Bool{Null: true}
	}

	if value, exists := object.GetLogVerbosity(); exists {
		state.LogVerbosity = types.Int64{Value: int64(value)}
	} else {
		state.LogVerbosity = types.Int64{Null: true}
	}

	if value, exists := object.GetMaxPodGracePeriod(); exists {
		state.MaxPodGracePeriod = types.Int64{Value: int64(value)}
	} else {
		state.MaxPodGracePeriod = types.Int64{Null: true}
	}

	if value, exists := object.GetPodPriorityThreshold(); exists {
		state.PodPriorityThreshold = types.Int64{Value: int64(value)}
	} else {
		state.PodPriorityThreshold = types.Int64{Null: true}
	}

	if value, exists := object.GetIgnoreDaemonsetsUtilization(); exists {
		state.IgnoreDaemonsetsUtilization = types.Bool{Value: value}
	} else {
		state.IgnoreDaemonsetsUtilization = types.Bool{Null: true}
	}

	state.MaxNodeProvisionTime = common.EmptiableStringToStringType(object.MaxNodeProvisionTime())

	if value, exists := object.GetBalancingIgnoredLabels(); exists {
		state.BalancingIgnoredLabels = common.StringArrayToList(value)
	} else {
		state.BalancingIgnoredLabels = types.List{Null: true}
	}

	if object.ResourceLimits() != nil {
		state.ResourceLimits = &AutoscalerResourceLimits{}

		if value, exists := object.ResourceLimits().GetMaxNodesTotal(); exists {
			state.ResourceLimits.MaxNodesTotal = types.Int64{Value: int64(value)}
		} else {
			state.ResourceLimits.MaxNodesTotal = types.Int64{Null: true}
		}

		cores := object.ResourceLimits().Cores()
		if cores != nil {
			state.ResourceLimits.Cores = &AutoscalerResourceRange{
				Min: types.Int64{Value: int64(cores.Min())},
				Max: types.Int64{Value: int64(cores.Max())},
			}
		}

		memory := object.ResourceLimits().Memory()
		if memory != nil {
			state.ResourceLimits.Memory = &AutoscalerResourceRange{
				Min: types.Int64{Value: int64(memory.Min())},
				Max: types.Int64{Value: int64(memory.Max())},
			}
		}

		gpus := object.ResourceLimits().GPUS()
		if gpus != nil {
			state.ResourceLimits.GPUS = make([]AutoscalerGPULimit, 0)

			for _, gpu := range gpus {
				state.ResourceLimits.GPUS = append(
					state.ResourceLimits.GPUS,
					AutoscalerGPULimit{
						Type: types.String{Value: gpu.Type()},
						Range: AutoscalerResourceRange{
							Min: types.Int64{Value: int64(gpu.Range().Min())},
							Max: types.Int64{Value: int64(gpu.Range().Max())},
						},
					},
				)
			}
		}
	}

	if object.ScaleDown() != nil {
		state.ScaleDown = &AutoscalerScaleDownConfig{}

		if value, exists := object.ScaleDown().GetEnabled(); exists {
			state.ScaleDown.Enabled = types.Bool{Value: value}
		} else {
			state.ScaleDown.Enabled = types.Bool{Null: true}
		}

		state.ScaleDown.UnneededTime = common.EmptiableStringToStringType(object.ScaleDown().UnneededTime())
		state.ScaleDown.UtilizationThreshold = common.EmptiableStringToStringType(object.ScaleDown().UtilizationThreshold())
		state.ScaleDown.DelayAfterAdd = common.EmptiableStringToStringType(object.ScaleDown().DelayAfterAdd())
		state.ScaleDown.DelayAfterDelete = common.EmptiableStringToStringType(object.ScaleDown().DelayAfterDelete())
		state.ScaleDown.DelayAfterFailure = common.EmptiableStringToStringType(object.ScaleDown().DelayAfterFailure())
	}
}

// clusterAutoscalerStateToObject builds a cluster-autoscaler API object from a given Terraform state.
func clusterAutoscalerStateToObject(state *ClusterAutoscalerState) (*cmv1.ClusterAutoscaler, error) {
	builder := cmv1.NewClusterAutoscaler()

	if !state.BalanceSimilarNodeGroups.Null {
		builder.BalanceSimilarNodeGroups(state.BalanceSimilarNodeGroups.Value)
	}

	if !state.SkipNodesWithLocalStorage.Null {
		builder.SkipNodesWithLocalStorage(state.SkipNodesWithLocalStorage.Value)
	}

	if !state.LogVerbosity.Null {
		builder.LogVerbosity(int(state.LogVerbosity.Value))
	}

	if !state.MaxPodGracePeriod.Null {
		builder.MaxPodGracePeriod(int(state.MaxPodGracePeriod.Value))
	}

	if !state.PodPriorityThreshold.Null {
		builder.PodPriorityThreshold(int(state.PodPriorityThreshold.Value))
	}

	if !state.IgnoreDaemonsetsUtilization.Null {
		builder.IgnoreDaemonsetsUtilization(state.IgnoreDaemonsetsUtilization.Value)
	}

	if !state.MaxNodeProvisionTime.Null {
		builder.MaxNodeProvisionTime(state.MaxNodeProvisionTime.Value)
	}

	if !state.BalancingIgnoredLabels.Null {
		builder.BalancingIgnoredLabels(common.OptionalList(state.BalancingIgnoredLabels)...)
	}

	if state.ResourceLimits != nil {
		resourceLimitsBuilder := cmv1.NewAutoscalerResourceLimits()

		if !state.ResourceLimits.MaxNodesTotal.Null {
			resourceLimitsBuilder.MaxNodesTotal(int(state.ResourceLimits.MaxNodesTotal.Value))
		}

		if state.ResourceLimits.Cores != nil {
			resourceLimitsBuilder.Cores(
				cmv1.NewResourceRange().
					Min(int(state.ResourceLimits.Cores.Min.Value)).
					Max(int(state.ResourceLimits.Cores.Max.Value)),
			)
		}

		if state.ResourceLimits.Memory != nil {
			resourceLimitsBuilder.Memory(
				cmv1.NewResourceRange().
					Min(int(state.ResourceLimits.Memory.Min.Value)).
					Max(int(state.ResourceLimits.Memory.Max.Value)),
			)
		}

		gpus := make([]*cmv1.AutoscalerResourceLimitsGPULimitBuilder, 0)
		for _, gpu := range state.ResourceLimits.GPUS {
			gpus = append(
				gpus,
				cmv1.NewAutoscalerResourceLimitsGPULimit().
					Type(gpu.Type.Value).
					Range(cmv1.NewResourceRange().
						Min(int(gpu.Range.Min.Value)).
						Max(int(gpu.Range.Max.Value))),
			)
		}
		resourceLimitsBuilder.GPUS(gpus...)

		builder.ResourceLimits(resourceLimitsBuilder)
	}

	if state.ScaleDown != nil {
		scaleDownBuilder := cmv1.NewAutoscalerScaleDownConfig()

		if !state.ScaleDown.Enabled.Null {
			scaleDownBuilder.Enabled(state.ScaleDown.Enabled.Value)
		}

		if !state.ScaleDown.UnneededTime.Null {
			scaleDownBuilder.UnneededTime(state.ScaleDown.UnneededTime.Value)
		}

		if !state.ScaleDown.UtilizationThreshold.Null {
			scaleDownBuilder.UtilizationThreshold(state.ScaleDown.UtilizationThreshold.Value)
		}

		if !state.ScaleDown.DelayAfterAdd.Null {
			scaleDownBuilder.DelayAfterAdd(state.ScaleDown.DelayAfterAdd.Value)
		}

		if !state.ScaleDown.DelayAfterDelete.Null {
			scaleDownBuilder.DelayAfterDelete(state.ScaleDown.DelayAfterDelete.Value)
		}

		if !state.ScaleDown.DelayAfterFailure.Null {
			scaleDownBuilder.DelayAfterFailure(state.ScaleDown.DelayAfterFailure.Value)
		}

		builder.ScaleDown(scaleDownBuilder)
	}

	return builder.Build()
}
