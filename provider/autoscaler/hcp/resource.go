/*
Copyright (c) 2024 Red Hat, Inc.

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

package hcp

import (
	"context"
	"fmt"
	"net/http"

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
	resp.TypeName = req.ProviderTypeName + "_hcp_cluster_autoscaler"
}

func (r *ClusterAutoscalerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Cluster-wide autoscaling configuration.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster." + common.ValueCannotBeChangedStringDescription,
				Required:    true,
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
			"max_node_provision_time": schema.StringAttribute{
				Description: "Maximum time cluster-autoscaler waits for node to be provisioned.",
				Optional:    true,
				Validators:  []validator.String{autoscaler.DurationStringValidator("max node provision time validation")},
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

	state.MaxNodeProvisionTime = common.EmptiableStringToStringType(object.MaxNodeProvisionTime())

	if object.ResourceLimits() != nil {
		state.ResourceLimits = &AutoscalerResourceLimits{}

		if value, exists := object.ResourceLimits().GetMaxNodesTotal(); exists {
			state.ResourceLimits.MaxNodesTotal = types.Int64Value(int64(value))
		} else {
			state.ResourceLimits.MaxNodesTotal = types.Int64Null()
		}
	}
	return nil
}

// clusterAutoscalerStateToObject builds a cluster-autoscaler API object from a given Terraform state.
func clusterAutoscalerStateToObject(state *ClusterAutoscalerState) (*cmv1.ClusterAutoscaler, error) {
	builder := cmv1.NewClusterAutoscaler()

	if !state.MaxPodGracePeriod.IsNull() {
		builder.MaxPodGracePeriod(int(state.MaxPodGracePeriod.ValueInt64()))
	}

	if !state.PodPriorityThreshold.IsNull() {
		builder.PodPriorityThreshold(int(state.PodPriorityThreshold.ValueInt64()))
	}

	if !state.MaxNodeProvisionTime.IsNull() {
		builder.MaxNodeProvisionTime(state.MaxNodeProvisionTime.ValueString())
	}

	if state.ResourceLimits != nil {
		resourceLimitsBuilder := cmv1.NewAutoscalerResourceLimits()

		if !state.ResourceLimits.MaxNodesTotal.IsNull() {
			resourceLimitsBuilder.MaxNodesTotal(int(state.ResourceLimits.MaxNodesTotal.ValueInt64()))
		}

		builder.ResourceLimits(resourceLimitsBuilder)
	}

	return builder.Build()
}
