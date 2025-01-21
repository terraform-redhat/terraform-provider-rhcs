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
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	failedToReadSummary   = "Failed to read KubeletConfig"
)

var createMutexKV = common.NewMutexKV()

type KubeletConfigResource struct {
	clusterClient common.ClusterClient
	configsClient client.KubeletConfigsClient
	clusterWait   common.ClusterWait
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
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the KubeletConfig." + common.ValueCannotBeChangedStringDescription,
			},
			"cluster": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the cluster." + common.ValueCannotBeChangedStringDescription,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
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
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the KubeletConfig." + common.ValueCannotBeChangedStringDescription,
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

	createMutexKV.Lock(clusterId)
	defer createMutexKV.Unlock(clusterId)

	isHCP, err := isHCP(ctx, clusterId, k.clusterClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't check cluster is HyperShift or not",
			err.Error(),
		)
		return
	}
	waitTimeoutInMinutes := int64(60)
	if _, err := k.clusterWait.WaitForClusterToBeReady(ctx, clusterId, waitTimeoutInMinutes); err != nil {
		resp.Diagnostics.AddError(
			"Cluster is not ready",
			fmt.Sprintf("Cluster with id '%s' is not in the ready state: %v", clusterId, err),
		)
		return
	}

	if !isHCP {
		configs, _, err := k.configsClient.List(ctx, clusterId, client.NewPaging(1, -1))
		if err != nil {
			resp.Diagnostics.AddError(failedToCreateSummary, err.Error())
			return
		}
		if len(configs) > 0 {
			resp.Diagnostics.AddError(failedToCreateSummary,
				fmt.Sprintf("KubeletConfig for cluster '%s' already exist", clusterId))
			return
		}
	}

	kubeletConfig, err := k.convertStateToApiResource(plan)
	if err != nil {
		resp.Diagnostics.AddError(failedToCreateSummary,
			fmt.Sprintf("Failed to build request to create KubeletConfig on cluster '%s': %v", clusterId, err))
		return
	}

	createdConfig, err := k.configsClient.Create(ctx, clusterId, kubeletConfig)
	if err != nil {
		resp.Diagnostics.AddError(failedToCreateSummary,
			fmt.Sprintf("Failed to create KubeletConfig on cluster '%s': %v", clusterId, err))
		return
	}

	plan.ID = types.StringValue(createdConfig.ID())
	plan.Name = types.StringValue(createdConfig.Name())
	k.writeStateToResponse(ctx, plan, &resp.State, &resp.Diagnostics)
}

func (k *KubeletConfigResource) convertStateToApiResource(state *KubeletConfigState) (*cmv1.KubeletConfig, error) {
	builder := cmv1.KubeletConfigBuilder{}
	builder.PodPidsLimit(int(state.PodPidsLimit.ValueInt64()))
	if !common.IsStringAttributeUnknownOrEmpty(state.Name) {
		builder.Name(state.Name.ValueString())
	}
	if !common.IsStringAttributeUnknownOrEmpty(state.ID) {
		builder.ID(state.ID.ValueString())
	}
	return builder.Build()
}

func (k *KubeletConfigResource) convertApiResourceToState(
	kubeletConfig *cmv1.KubeletConfig, state *KubeletConfigState) {
	state.PodPidsLimit = types.Int64Value(int64(kubeletConfig.PodPidsLimit()))
	state.Name = types.StringValue(kubeletConfig.Name())
	state.ID = types.StringValue(kubeletConfig.ID())
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
	name := state.Name.ValueString()
	kubeletConfigId, err := getKubeletConfigId(ctx, state, clusterId, name, k.configsClient)
	if err != nil {
		resp.Diagnostics.AddError(failedToReadSummary, err.Error())
		return
	}

	exists, kubeletConfig, err := k.configsClient.Exists(ctx, clusterId, kubeletConfigId)
	if err != nil {
		resp.Diagnostics.AddError(failedToReadSummary,
			fmt.Sprintf("Cannot read KubeletConfig '%s' for cluster '%s': %v", kubeletConfigId, clusterId, err))
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
	state := k.getKubeletConfigStateFromState(ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := k.getKubeletConfigStateFromPlan(ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// assert attribute cluster&name weren't changed:
	common.ValidateStateAndPlanEquals(state.Cluster, plan.Cluster, "cluster", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.IsStringAttributeUnknownOrEmpty(plan.Name) {
		common.ValidateStateAndPlanEquals(state.Name, plan.Name, "name", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	clusterId := state.Cluster.ValueString()
	name := state.Name.ValueString()
	kubeletConfigId, err := getKubeletConfigId(ctx, state, clusterId, name, k.configsClient)
	if err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary, err.Error())
		return
	}
	if _, err := k.configsClient.Get(ctx, clusterId, kubeletConfigId); err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary,
			fmt.Sprintf("Failed to read details of existing KubeletConfig '%s' for cluster '%s': %v", kubeletConfigId, clusterId, err))
		return
	}

	plan.ID = types.StringValue(kubeletConfigId)
	kubeletConfig, err := k.convertStateToApiResource(plan)
	if err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary,
			fmt.Sprintf("Failed to build request to update existing KubeletConfig '%s' for cluster '%s': %v",
				kubeletConfigId, clusterId, err))
		return
	}

	updateResponse, err := k.configsClient.Update(ctx, clusterId, kubeletConfig)
	if err != nil {
		resp.Diagnostics.AddError(failedToUpdateSummary,
			fmt.Sprintf("Failed to update existing KubeletConfig '%s' for cluster '%s': %v",
				kubeletConfigId, clusterId, err))
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
	name := state.Name.ValueString()
	kubeletConfigId, err := getKubeletConfigId(ctx, state, clusterId, name, k.configsClient)
	if err != nil {
		resp.Diagnostics.AddError(failedToDeleteSummary, err.Error())
		return
	}

	err = k.configsClient.Delete(ctx, clusterId, kubeletConfigId)
	if err != nil {
		resp.Diagnostics.AddError(failedToDeleteSummary,
			fmt.Sprintf("Failed to delete KubeletConfig '%s' for cluster '%s': %v", kubeletConfigId, clusterId, err))
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

	clusterCollection := connection.ClustersMgmt().V1().Clusters()
	k.clusterClient = common.NewClusterClient(clusterCollection)
	k.configsClient = client.NewKubeletConfigsClient(clusterCollection)
	k.clusterWait = common.NewClusterWait(clusterCollection, connection)
}

func isHCP(ctx context.Context, clusterId string, clusterClient common.ClusterClient) (bool, error) {
	isHCP := false
	cluster, err := clusterClient.FetchCluster(ctx, clusterId)
	if err != nil {
		return isHCP, err
	}
	return cluster.Hypershift().Enabled(), nil
}

func getKubeletConfigId(ctx context.Context, state *KubeletConfigState, clusterId string, name string,
	configsClient client.KubeletConfigsClient) (string, error) {
	id := ""
	if common.IsStringAttributeUnknownOrEmpty(state.ID) {
		kubeletConfigs, _, err := configsClient.List(ctx, clusterId, client.NewPaging(1, -1))
		if err != nil {
			return id, fmt.Errorf("Cannot list KubeletConfigs for cluster '%s': %v", clusterId, err)
		}
		// for classic cluster, there is maxium one kubeletconfig
		if name == "" {
			if len(kubeletConfigs) > 0 {
				return kubeletConfigs[0].ID(), nil
			} else {
				return id, fmt.Errorf("Cannot find KubeletConfig for cluster '%s'", clusterId)
			}
		}
		for _, config := range kubeletConfigs {
			if config.Name() == name {
				id = config.ID()
				break
			}
		}
		if id == "" {
			return id, fmt.Errorf("Cannot find KubeletConfig '%s' for cluster '%s'", name, clusterId)
		}
	} else {
		id = state.ID.ValueString()
	}
	return id, nil
}
