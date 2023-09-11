/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"context"
***REMOVED***
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocmr "github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
***REMOVED***

type DNSDomainResource struct {
	collection *cmv1.DNSDomainsClient
}

var _ resource.ResourceWithConfigure = &DNSDomainResource{}
var _ resource.ResourceWithImportState = &DNSDomainResource{}

func New(***REMOVED*** resource.Resource {
	return &DNSDomainResource{}
}

func (r *DNSDomainResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_dns_domain"
}

func (r *DNSDomainResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "DNS Domain",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the DNS Domain",
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
}

func (r *DNSDomainResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection***REMOVED***
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData***REMOVED***,
		***REMOVED***
		return
	}

	r.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.DNSDomains(***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse***REMOVED*** {
	// Get the current state
	state := &DNSDomainState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the DNS domain
	dnsDomain := ocmr.NewDNSDomain(r.collection***REMOVED***
	get, err := dnsDomain.Get(state.ID.ValueString(***REMOVED******REMOVED***
	if err != nil && get.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, "DNS domain not found, removing from state", map[string]interface{}{
			"id": state.ID.ValueString(***REMOVED***,
***REMOVED******REMOVED***
		resp.State.RemoveResource(ctx***REMOVED***
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf(
				"Can't find DNS domain with identifier '%s'",
				state.ID.ValueString(***REMOVED***,
			***REMOVED***,
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}

	object := get.Body(***REMOVED***

	if id, ok := object.GetID(***REMOVED***; ok {
		state.ID = types.StringValue(id***REMOVED***
	}

	// Save the state
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse***REMOVED*** {
	// Get the plan
	plan := &DNSDomainState{}
	diags := req.Plan.Get(ctx, plan***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	dnsDomain := ocmr.NewDNSDomain(r.collection***REMOVED***
	createResp, err := dnsDomain.Create(***REMOVED***

	if err != nil {
		resp.Diagnostics.AddError("Failed to create DNS Domain.", err.Error(***REMOVED******REMOVED***
		return
	}

	if id, ok := createResp.Body(***REMOVED***.GetID(***REMOVED***; ok {
		plan.ID = types.StringValue(id***REMOVED***
	} else {
		resp.Diagnostics.AddError("Failed to create DNS Domain.", "Failed to get an ID"***REMOVED***
		return
	}

	// Save the state
	diags = resp.State.Set(ctx, plan***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse***REMOVED*** {
	// Until we support. return an informative error
	resp.Diagnostics.AddError("Can't update DNS domain", "Update is currently not supported."***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse***REMOVED*** {
	// Get the state
	state := &DNSDomainState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	dnsDomain := ocmr.NewDNSDomain(r.collection***REMOVED***
	err := dnsDomain.Delete(state.ID.ValueString(***REMOVED******REMOVED***

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Can't delete DNS domain with identifier '%s'", state.ID.ValueString(***REMOVED******REMOVED***,
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state
	resp.State.RemoveResource(ctx***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse***REMOVED*** {
	resource.ImportStatePassthroughID(ctx, path.Root("id"***REMOVED***, req, resp***REMOVED***
}
