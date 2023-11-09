/*
Copyright (c) 2021 Red Hat, Inc.

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

package dnsdomain

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocmr "github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
)

type DNSDomainResource struct {
	collection *cmv1.DNSDomainsClient
}

var _ resource.ResourceWithConfigure = &DNSDomainResource{}
var _ resource.ResourceWithImportState = &DNSDomainResource{}

func New() resource.Resource {
	return &DNSDomainResource{}
}

func (r *DNSDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_domain"
}

func (r *DNSDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "DNS Domain",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the DNS Domain",
				Computed:    true,
			},
		},
	}
}

func (r *DNSDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.collection = connection.ClustersMgmt().V1().DNSDomains()
}

func (r *DNSDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	state := &DNSDomainState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Find the DNS domain
	dnsDomain := ocmr.NewDNSDomain(r.collection)
	get, err := dnsDomain.Get(state.ID.ValueString())
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, "DNS domain not found, removing from state", map[string]interface{}{
			"id": state.ID.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf(
				"Can't find DNS domain with identifier '%s'",
				state.ID.ValueString(),
			),
			err.Error(),
		)
		return
	}

	object := get.Body()

	if id, ok := object.GetID(); ok {
		state.ID = types.StringValue(id)
	}

	// Save the state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DNSDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan
	plan := &DNSDomainState{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnsDomain := ocmr.NewDNSDomain(r.collection)
	createResp, err := dnsDomain.Create()

	if err != nil {
		resp.Diagnostics.AddError("Failed to create DNS Domain.", err.Error())
		return
	}

	if id, ok := createResp.Body().GetID(); ok {
		plan.ID = types.StringValue(id)
	} else {
		resp.Diagnostics.AddError("Failed to create DNS Domain.", "Failed to get an ID")
		return
	}

	// Save the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DNSDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Until we support. return an informative error
	resp.Diagnostics.AddError("Can't update DNS domain", "Update is currently not supported.")
}

func (r *DNSDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the state
	state := &DNSDomainState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnsDomain := ocmr.NewDNSDomain(r.collection)
	err := dnsDomain.Delete(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Can't delete DNS domain with identifier '%s'", state.ID.ValueString()),
			err.Error(),
		)
		return
	}

	// Remove the state
	resp.State.RemoveResource(ctx)
}

func (r *DNSDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
