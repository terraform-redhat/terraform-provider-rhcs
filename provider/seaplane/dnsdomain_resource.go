// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/providerdata"
)

var (
	_ resource.Resource                = (*DNSDomainSeaplaneResource)(nil)
	_ resource.ResourceWithConfigure   = (*DNSDomainSeaplaneResource)(nil)
	_ resource.ResourceWithImportState = (*DNSDomainSeaplaneResource)(nil)
)

// DNSDomainSeaplaneResource manages a Seaplane DNS reservation.
type DNSDomainSeaplaneResource struct {
	client *DNSDomainSeaplaneClient
}

func (r *DNSDomainSeaplaneResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_dns_domain_seaplane"
}

func (r *DNSDomainSeaplaneResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an account-scoped DNS reservation through the Seaplane Platform API. " +
			"A reservation can be deleted only after the cluster using it has been destroyed.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Customer-selected reservation name. Changing this value creates a new reservation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Reservation UID assigned by the Seaplane Platform API. " +
					"Use this value for `rhcs_cluster_seaplane.dns_reservation_id`.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"base_domain": schema.StringAttribute{
				MarkdownDescription: "DNS base domain assigned to the reservation, such as `f7a3.0.openshiftapps.com`.",
				Computed:            true,
			},
			"phase": schema.StringAttribute{
				MarkdownDescription: "Current reservation lifecycle phase. Terraform waits for `Ready` before completing create.",
				Computed:            true,
			},
		},
	}
}

func (r *DNSDomainSeaplaneResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	shared, ok := req.ProviderData.(*providerdata.ProviderSharedData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *providerdata.ProviderSharedData, got: %T", req.ProviderData),
		)
		return
	}
	if shared.SeaplaneClient == nil {
		resp.Diagnostics.AddError(
			"Seaplane not configured",
			"The seaplane_url provider attribute is required to manage DNS reservation resources. "+
				"Set it explicitly in the provider configuration or via the RHCS_SEAPLANE_URL environment variable.",
		)
		return
	}

	r.client = NewDNSDomainSeaplaneClient(shared.SeaplaneClient)
}

func (r *DNSDomainSeaplaneResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	state := &DNSDomainSeaplaneState{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reservation, err := r.client.Get(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to read DNS reservation %q", state.Name.ValueString()),
			err.Error(),
		)
		return
	}
	if reservation == nil {
		tflog.Warn(ctx, "DNS reservation not found, removing from state", map[string]any{
			"name": state.Name.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateDNSDomainSeaplaneState(reservation, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *DNSDomainSeaplaneResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	plan := &DNSDomainSeaplaneState{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reservation, err := r.client.Create(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create DNS reservation", err.Error())
		return
	}

	populateDNSDomainSeaplaneState(reservation, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DNSDomainSeaplaneResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError(
		"Can't update DNS reservation",
		"Reservation names are immutable; changing name replaces the reservation.",
	)
}

func (r *DNSDomainSeaplaneResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	state := &DNSDomainSeaplaneState{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to delete DNS reservation %q", state.Name.ValueString()),
			err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *DNSDomainSeaplaneResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	reservation, err := r.client.Get(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import DNS reservation", err.Error())
		return
	}
	if reservation == nil {
		resp.Diagnostics.AddError(
			"DNS reservation not found",
			fmt.Sprintf("No DNS reservation named %q was found.", req.ID),
		)
		return
	}

	state := &DNSDomainSeaplaneState{}
	populateDNSDomainSeaplaneState(reservation, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func populateDNSDomainSeaplaneState(reservation *DNSDomainSeaplane, state *DNSDomainSeaplaneState) {
	state.Name = types.StringValue(reservation.Name)
	state.ID = types.StringValue(reservation.ID)
	state.BaseDomain = types.StringValue(reservation.BaseDomain)
	state.Phase = types.StringValue(reservation.Phase)
}
