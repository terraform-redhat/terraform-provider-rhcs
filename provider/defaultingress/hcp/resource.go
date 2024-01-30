package hcp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

var validListeningMethods = []string{string(cmv1.ListeningMethodExternal), string(cmv1.ListeningMethodInternal)}

type DefaultIngressResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

func New() resource.Resource {
	return &DefaultIngressResource{}
}

var _ resource.Resource = &DefaultIngressResource{}
var _ resource.ResourceWithImportState = &DefaultIngressResource{}
var _ resource.ResourceWithConfigure = &DefaultIngressResource{}

func (r *DefaultIngressResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hcp_default_ingress"
}

func (r *DefaultIngressResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Edit a cluster ingress (load balancer)",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the ingress.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					// This passes the state through to the plan, preventing
					// "known after apply" since we know it won't change.
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"listening_method": schema.StringAttribute{
				Description: fmt.Sprintf("Listening Method for apps ingress. Options are %s.",
					strings.Join(validListeningMethods, ",")),
				Required:   true,
				Validators: []validator.String{attrvalidators.EnumValueValidator(validListeningMethods)},
			},
		},
	}
	return
}

func (r *DefaultIngressResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DefaultIngressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &DefaultIngress{}
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
	err = r.updateIngress(ctx, nil, plan, plan.Cluster.ValueString(), r.collection)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed building cluster default ingress",
			fmt.Sprintf(
				"Failed building default ingress for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DefaultIngressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &DefaultIngress{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.populateDefaultIngress(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed getting cluster default ingress",
			fmt.Sprintf(
				"Failed getting default ingress for cluster '%s': %v",
				state.Cluster.ValueString(), err,
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

func (r *DefaultIngressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get the state:
	state := &DefaultIngress{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &DefaultIngress{}
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

	err := r.updateIngress(ctx, state, plan, plan.Cluster.ValueString(), r.collection)
	if err != nil {
		diags.AddError(
			"Failed to update default ingress",
			fmt.Sprintf(
				"Cannot update default ingress for "+
					"cluster '%s': %v", state.Cluster.ValueString(), err,
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

func (r *DefaultIngressResource) Delete(ctx context.Context, req resource.DeleteRequest,
	resp *resource.DeleteResponse) {
	// Until we support. return an informative error
	state := &DefaultIngress{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.AddWarning(
		"Cannot delete default ingress",
		fmt.Sprintf(
			"Cannot delete default ingress for cluster '%s'. "+
				"ROSA Classic clusters must have default ingress. "+
				"It is being removed from the Terraform state only. "+
				"To resume managing default ingress, import it again. "+
				"It will be automatically deleted when the cluster is deleted.",
			state.Cluster.ValueString(),
		),
	)
	// Remove the state:
	resp.State.RemoveResource(ctx)

}

func (r *DefaultIngressResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	tflog.Debug(ctx, "begin importstate()")

	resource.ImportStatePassthroughID(ctx, path.Root("cluster"), request, response)
}

func (r *DefaultIngressResource) populateDefaultIngress(
	ctx context.Context,
	state *DefaultIngress) error {

	var ingress *cmv1.Ingress
	var err error
	if common.IsStringAttributeUnknownOrEmpty(state.Id) {
		ingress, err = r.populateDefaultIngressFromList(ctx, state)
		if err != nil {
			return err
		}
	} else {
		ingressResp, err := r.collection.Cluster(state.Cluster.ValueString()).Ingresses().Ingress(state.Id.ValueString()).Get().SendContext(ctx)
		if err != nil {
			return err
		}
		ingress = ingressResp.Body()
	}

	return r.populateState(ingress, state)
}

func (r *DefaultIngressResource) populateDefaultIngressFromList(ctx context.Context, state *DefaultIngress) (*cmv1.Ingress, error) {
	ingresses, err := r.collection.Cluster(state.Cluster.ValueString()).Ingresses().List().SendContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, ingress := range ingresses.Items().Slice() {
		if ingress.Default() {
			return ingress, nil
		}
	}

	return nil, fmt.Errorf("failed to find default ingress")
}

func (r *DefaultIngressResource) populateState(ingress *cmv1.Ingress, state *DefaultIngress) error {
	if state == nil {
		state = &DefaultIngress{}
	}
	state.Id = types.StringValue(ingress.ID())
	state.ListeningMethod = types.StringValue(string(ingress.Listening()))

	return nil
}

func (r *DefaultIngressResource) updateIngress(ctx context.Context, state, plan *DefaultIngress,
	clusterId string, clusterCollection *cmv1.ClustersClient) error {

	if state == nil {
		state = &DefaultIngress{Cluster: plan.Cluster}
	}
	// In case default ingress was not part of state till now and we want to set
	// it we need to bring it first as we need to set specific id
	if common.IsStringAttributeUnknownOrEmpty(state.Id) {
		err := r.populateDefaultIngress(ctx, state)
		if err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(state, plan) {
		err := validateDefaultIngress(ctx, plan)
		if err != nil {
			return err
		}
		if plan == nil {
			plan = &DefaultIngress{}
		}

		ingressBuilder := getDefaultIngressBuilder(ctx, state, plan)

		ingress, err := ingressBuilder.Build()
		if err != nil {
			return err
		}

		ingressResp, err := clusterCollection.Cluster(clusterId).Ingresses().Ingress(state.Id.ValueString()).Update().
			Body(ingress).SendContext(ctx)

		if err != nil {
			return err
		}
		if err := r.populateState(ingressResp.Body(), plan); err != nil {
			return err
		}
	}

	return nil
}

func getDefaultIngressBuilder(ctx context.Context, state, plan *DefaultIngress) *cmv1.IngressBuilder {
	ingressBuilder := cmv1.NewIngress()
	if !common.IsStringAttributeUnknownOrEmpty(plan.ListeningMethod) && state.ListeningMethod != plan.ListeningMethod {
		ingressBuilder.Listening(cmv1.ListeningMethod(plan.ListeningMethod.ValueString()))
	}
	return ingressBuilder
}

func validateDefaultIngress(ctx context.Context, state *DefaultIngress) error {
	return nil
}
