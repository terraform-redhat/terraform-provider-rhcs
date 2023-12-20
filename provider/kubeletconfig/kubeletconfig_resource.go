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

package kubeletconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/openshift-online/ocm-common/pkg/ocm/client"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

const (
	resourceTypeName      = "_kubeletconfig"
	failedToCreateSummary = "Failed to create KubeletConfig"
	failedToUpdateSummary = "Failed to update KubeletConfig"
	failedToDeleteSummary = "Failed to delete KubeletConfig"
)

type KubeletConfigResource struct {
	configClient client.KubeletConfigClient
	clusterWait  common.ClusterWait
}

// Interface checks
var _ resource.Resource = &KubeletConfigResource{}
var _ resource.ResourceWithConfigure = &KubeletConfigResource{}
var _ resource.ResourceWithImportState = &KubeletConfigResource{}

func New() resource.Resource {
	return &KubeletConfigResource{}
}

func (k *KubeletConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("cluster"), request, response)
}

func (k *KubeletConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + resourceTypeName
}

func (k *KubeletConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "KubeletConfig allows setting a customized Kubelet configuration",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the cluster.",
				PlanModifiers: []planmodifier.String{
					common.Immutable(),
				},
			},
			"pod_pids_limit": schema.Int64Attribute{
				Required: true,
				Description: "Sets the requested podPidsLimit to be applied as part of the custom " +
					"KubeletConfig.",
				Validators: []validator.Int64{
					PidsLimitValidator{},
				},
			},
		},
	}
}

func (k *KubeletConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	plan := k.getKubeletConfigStateFromPlan(ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.Cluster.ValueString()

	if err := k.clusterWait.WaitForClusterToBeReady(ctx, clusterId); err != nil {
		resp.Diagnostics.AddError(
			"Cluster is not ready",
			fmt.Sprintf("Cluster with id '%s' is not in the ready state: %v", clusterId, err),
		)
		return
	}

	if exists, _, err := k.configClient.Exists(ctx, clusterId); err != nil || exists {
		resp.Diagnostics.AddError(failedToCreateSummary,
			fmt.Sprintf("KubeletConfig for cluster '%s' may already exist: %v", clusterId, err))
		return
	}

	kubeletConfig, err := k.convertStateToApiResource(plan)
	if err != nil {
		resp.Diagnostics.AddError(failedToCreateSummary,
			fmt.Sprintf("Failed to build request to create KubeletConfig on cluster '%s': %v", clusterId, err))
		return
	}

	if _, err = k.configClient.Create(ctx, clusterId, kubeletConfig); err != nil {
		resp.Diagnostics.AddError(failedToCreateSummary,
			fmt.Sprintf("Failed to create KubeletConfig on cluster '%s': %v", clusterId, err))
		return
	}

	k.writeStateToResponse(ctx, plan, &resp.State, &resp.Diagnostics)
}

func (k *KubeletConfigResource) convertStateToApiResource(state *KubeletConfigState) (*cmv1.KubeletConfig, error) {
	builder := cmv1.KubeletConfigBuilder{}
	builder.PodPidsLimit(int(state.PodPidsLimit.ValueInt64()))
	return builder.Build()
}

func (k *KubeletConfigResource) convertApiResourceToState(
	kubeletConfig *cmv1.KubeletConfig, state *KubeletConfigState) {
	state.PodPidsLimit = types.Int64Value(int64(kubeletConfig.PodPidsLimit()))
}

func (k *KubeletConfigResource) getKubeletConfigStateFromState(
	ctx context.Context, state tfsdk.State, diags *diag.Diagnostics) *KubeletConfigState {
	ks := &KubeletConfigState{}
	ds := state.Get(ctx, ks)
	diags.Append(ds...)
	return ks
}

func (k *KubeletConfigResource) getKubeletConfigStateFromPlan(
	ctx context.Context, plan tfsdk.Plan, diags *diag.Diagnostics) *KubeletConfigState {
	ks := &KubeletConfigState{}
	ds := plan.Get(ctx, ks)
	diags.Append(ds...)
	return ks
}

func (k *KubeletConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := k.getKubeletConfigStateFromState(ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.Cluster.ValueString()
	exists, kubeletConfig, err := k.configClient.Exists(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError("Cannot read KubeletConfig",
			fmt.Sprintf("Cannot read KubeletConfig for cluster '%s': %v", clusterId, err))
		return
	}

	if !exists {
		tflog.Warn(ctx, fmt.Sprintf("kubeletconfig for cluster (%s) not found, removing from plan",
			state.Cluster.ValueString(),
		))
		resp.State.RemoveResource(ctx)
		return
	}

	k.convertApiResourceToState(kubeletConfig, state)
	k.writeStateToResponse(ctx, state, &resp.State, &resp.Diagnostics)
}

func (k *KubeletConfigResource) writeStateToResponse(
	ctx context.Context, ks *KubeletConfigState, tfState *tfsdk.State, diagnostics *diag.Diagnostics) {

	diags := tfState.Set(ctx, ks)
	diagnostics.Append(diags...)
}

func (k *KubeletConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	plan := k.getKubeletConfigStateFromPlan(ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.Cluster.ValueString()

	if _, err := k.configClient.Get(ctx, clusterId); err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary,
			fmt.Sprintf("Failed to read details of existing KubeletConfig for cluster '%s': %v", clusterId, err))
		return
	}

	kubeletConfig, err := k.convertStateToApiResource(plan)
	if err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary,
			fmt.Sprintf("Failed to build request to update existing KubeletConfig for cluster '%s': %v", clusterId, err))
		return
	}

	updateResponse, err := k.configClient.Update(ctx, clusterId, kubeletConfig)
	if err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary,
			fmt.Sprintf("Failed to update existing KubeletConfig for cluster '%s': %v", clusterId, err))
		return
	}

	k.convertApiResourceToState(updateResponse, plan)
	k.writeStateToResponse(ctx, plan, &resp.State, &resp.Diagnostics)
}

func (k *KubeletConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := k.getKubeletConfigStateFromState(ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.Cluster.ValueString()
	err := k.configClient.Delete(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError(failedToDeleteSummary,
			fmt.Sprintf("Failed to delete KubeletConfig for cluster '%s': %v", clusterId, err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (k *KubeletConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

	collection := connection.ClustersMgmt().V1().Clusters()
	k.configClient = client.NewKubeletConfigClient(collection)
	k.clusterWait = common.NewClusterWait(collection)
}
