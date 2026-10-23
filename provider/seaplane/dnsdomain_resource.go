// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// DNSDomainSeaplaneResource manages a Seaplane HCP DNS-domain reservation.
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
		MarkdownDescription: "Manages an account-scoped HCP DNS-domain reservation through the Seaplane Platform API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Reserved DNS domain ID, which is also the base domain in the reservation.",
				Computed:            true,
			},
			"cluster_arch": schema.StringAttribute{
				MarkdownDescription: "Cluster architecture of the DNS domain. Seaplane currently supports only `hcp`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(dnsDomainClusterArchHCP),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(dnsDomainClusterArchHCP),
				},
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
			"The seaplane_url provider attribute is required to manage DNS-domain resources. "+
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

	domain, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to read DNS domain %q", state.ID.ValueString()),
			err.Error(),
		)
		return
	}
	if domain == nil {
		tflog.Warn(ctx, "DNS domain not found, removing from state", map[string]any{
			"id": state.ID.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateDNSDomainSeaplaneState(domain, state)
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

	domain, err := r.client.Create(ctx, plan.ClusterArch.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create DNS domain", err.Error())
		return
	}

	populateDNSDomainSeaplaneState(domain, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DNSDomainSeaplaneResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError("Can't update DNS domain", "Update is currently not supported.")
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

	if err := r.client.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to delete DNS domain %q", state.ID.ValueString()),
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func populateDNSDomainSeaplaneState(domain *DNSDomainSeaplane, state *DNSDomainSeaplaneState) {
	state.ID = types.StringValue(domain.ID)
	if domain.ClusterArch != "" {
		state.ClusterArch = types.StringValue(domain.ClusterArch)
	} else {
		state.ClusterArch = types.StringValue(dnsDomainClusterArchHCP)
	}
}
