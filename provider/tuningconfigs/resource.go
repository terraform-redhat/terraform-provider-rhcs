package tuningconfigs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type TuningConfigResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

func New() resource.Resource {
	return &TuningConfigResource{}
}

var _ resource.Resource = &TuningConfigResource{}
var _ resource.ResourceWithImportState = &TuningConfigResource{}
var _ resource.ResourceWithConfigure = &TuningConfigResource{}

func (r *TuningConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tuning_config"
}

func (r *TuningConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Edit a cluster tuning config",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the tuning config.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					// This passes the state through to the plan, preventing
					// "known after apply" since we know it won't change.
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the tuning configuration. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"spec": schema.StringAttribute{
				Description: "Definition of the spec",
				Required:    true,
			},
		},
	}
	return
}

func (r *TuningConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TuningConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &TuningConfig{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
	err = r.createTuningConfig(ctx, plan, plan.Cluster.ValueString(), r.collection)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed building cluster tuning config",
			fmt.Sprintf(
				"Failed building tuning config for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *TuningConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &TuningConfig{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.populateTuningConfig(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed getting cluster tuning config",
			fmt.Sprintf(
				"Failed getting tuning config with"+
					" identifier '%s' for cluster '%s': %v",
				state.Id.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *TuningConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get the state:
	state := &TuningConfig{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &TuningConfig{}
	diags = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// assert cluster attribute wasn't changed:
	common.ValidateStateAndPlanEquals(state.Cluster, plan.Cluster, "cluster", &diags)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateTuningConfig(ctx, state, plan, plan.Cluster.ValueString(), r.collection)
	if err != nil {
		diags.AddError(
			"Failed to update tuning config",
			fmt.Sprintf(
				"Failed to update tuning config with"+
					" identifier '%s' for cluster '%s': %v",
				state.Id.ValueString(), state.Cluster.ValueString(), err,
			),
		)
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *TuningConfigResource) Delete(ctx context.Context, req resource.DeleteRequest,
	resp *resource.DeleteResponse) {
	// Until we support. return an informative error
	state := &TuningConfig{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Send the request to delete the machine pool:
	resource := r.collection.Cluster(state.Cluster.ValueString()).
		TuningConfigs().
		TuningConfig(state.Id.ValueString())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete tuning config",
			fmt.Sprintf(
				"Cannot delete tuning config with identifier '%s' for "+
					"cluster '%s': %v",
				state.Id.ValueString(), state.Cluster.ValueString(), err,
			),
		)
	}
	// Remove the state:
	resp.State.RemoveResource(ctx)

}

func (r *TuningConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	tflog.Debug(ctx, "begin importstate()")

	resource.ImportStatePassthroughID(ctx, path.Root("cluster"), request, response)
}

func (r *TuningConfigResource) populateTuningConfig(
	ctx context.Context,
	state *TuningConfig) error {

	tuningConfigResp, err := r.collection.Cluster(state.Cluster.ValueString()).TuningConfigs().TuningConfig(state.Id.ValueString()).Get().SendContext(ctx)
	if err != nil {
		return err
	}
	tuningConfig := tuningConfigResp.Body()

	return r.populateState(tuningConfig, state)
}

func (r *TuningConfigResource) populateState(tuningConfig *cmv1.TuningConfig, state *TuningConfig) error {
	if state == nil {
		state = &TuningConfig{}
	}
	state.Id = types.StringValue(tuningConfig.ID())
	state.Name = types.StringValue(tuningConfig.Name())
	// Pretty print the spec
	tuningConfigSpec, err := json.Marshal(tuningConfig.Spec())
	if err != nil {
		return err
	}
	state.Spec = types.StringValue(string(tuningConfigSpec[:]))

	return nil
}

func (r *TuningConfigResource) createTuningConfig(ctx context.Context, plan *TuningConfig,
	clusterId string, clusterCollection *cmv1.ClustersClient) error {

	if plan == nil {
		plan = &TuningConfig{}
	}

	tuningConfigBuilder, err := getTuningConfigBuilder(ctx, plan)
	if err != nil {
		return err
	}
	tuningConfigBuilder.Name(plan.Name.ValueString())

	tuningConfig, err := tuningConfigBuilder.Build()
	if err != nil {
		return err
	}

	tuningConfigResp, err := clusterCollection.Cluster(clusterId).TuningConfigs().Add().
		Body(tuningConfig).SendContext(ctx)

	if err != nil {
		return err
	}
	if err := r.populateState(tuningConfigResp.Body(), plan); err != nil {
		return err
	}

	return nil
}

func (r *TuningConfigResource) updateTuningConfig(ctx context.Context, state, plan *TuningConfig,
	clusterId string, clusterCollection *cmv1.ClustersClient) error {

	if state == nil {
		state = &TuningConfig{Cluster: plan.Cluster}
	}
	if common.IsStringAttributeUnknownOrEmpty(state.Id) {
		err := r.populateTuningConfig(ctx, state)
		if err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(state, plan) {
		if plan == nil {
			plan = &TuningConfig{}
		}

		tuningConfigBuilder, err := getTuningConfigBuilder(ctx, plan)
		if err != nil {
			return err
		}

		tuningConfig, err := tuningConfigBuilder.Build()
		if err != nil {
			return err
		}

		tuningConfigResp, err := clusterCollection.Cluster(clusterId).TuningConfigs().TuningConfig(state.Id.ValueString()).Update().
			Body(tuningConfig).SendContext(ctx)

		if err != nil {
			return err
		}
		if err := r.populateState(tuningConfigResp.Body(), plan); err != nil {
			return err
		}
	}

	return nil
}

func getTuningConfigBuilder(ctx context.Context, plan *TuningConfig) (*cmv1.TuningConfigBuilder, error) {
	tuningConfigBuilder := cmv1.NewTuningConfig()
	if !common.IsStringAttributeUnknownOrEmpty(plan.Spec) {
		parsedSpec, err := parseInputString([]byte(plan.Spec.ValueString()))
		if err != nil {
			return nil, err
		}
		tuningConfigBuilder.Spec(parsedSpec)
	}
	return tuningConfigBuilder, nil
}

func parseInputString(input []byte) (map[string]interface{}, error) {
	var validSpecJson map[string]interface{}
	err := json.Unmarshal(input, &validSpecJson)
	if err != nil {
		return nil, err
	}
	return validSpecJson, nil
}
