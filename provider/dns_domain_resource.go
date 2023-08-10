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

package provider

***REMOVED***
	"context"
***REMOVED***
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
***REMOVED***

type DNSDomainResourceType struct {
}

type DNSDomainResource struct {
	collection *cmv1.DNSDomainsClient
}

func (t *DNSDomainResourceType***REMOVED*** NewResource(ctx context.Context, p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	// use it directly when needed.
	parent := p.(*Provider***REMOVED***

	//Create DNS Clients - not related to a cluster.
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.DNSDomains(***REMOVED***

	// Create the resource
	result = &DNSDomainResource{
		collection: collection,
	}

	return
}

func (t *DNSDomainResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "DNS Domain",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier of the DNS Domain",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}

	return
}

func (r *DNSDomainResource***REMOVED*** Read(ctx context.Context,
	request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse***REMOVED*** {

	// Get the current state
	state := &DNSDomainState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the identity provider
	dnsDomain := resource.NewDNSDomain(r.collection***REMOVED***
	get, err := dnsDomain.Get(state.ID.Value***REMOVED***
	if err != nil && get.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("DNS domain (%s***REMOVED*** not found, removing from state",
			state.ID.Value,
		***REMOVED******REMOVED***
		response.State.RemoveResource(ctx***REMOVED***
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(
				"Can't find DNS domain with identifier '%s'",
				state.ID.Value,
			***REMOVED***,
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}

	object := get.Body(***REMOVED***

	if id, ok := object.GetID(***REMOVED***; ok {
		state.ID = types.String{
			Value: id,
***REMOVED***
	}

	// Save the state
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {

	// Get the plan
	plan := &DNSDomainState{}
	diags := request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	dnsDomain := resource.NewDNSDomain(r.collection***REMOVED***
	resp, err := dnsDomain.Create(***REMOVED***

	if err != nil {
		response.Diagnostics.AddError("Failed to create DNS Domain.", err.Error(***REMOVED******REMOVED***
		return
	}

	if id, ok := resp.Body(***REMOVED***.GetID(***REMOVED***; ok {
		plan.ID = types.String{
			Value: id,
***REMOVED***
	} else {
		response.Diagnostics.AddError("Failed to create DNS Domain.", "Failed to get an ID"***REMOVED***
		return
	}

	// Save the state
	diags = response.State.Set(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Update(ctx context.Context,
	request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse***REMOVED*** {
	// Until we support. return an informative error
	response.Diagnostics.AddError("Can't update DNS domain", "Update is currently not supported."***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** Delete(ctx context.Context,
	request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse***REMOVED*** {
	// Get the state
	state := &DNSDomainState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	dnsDomain := resource.NewDNSDomain(r.collection***REMOVED***
	err := dnsDomain.Delete(state.ID.Value***REMOVED***

	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("Can't delete DNS domain with identifier '%s'", state.ID.Value***REMOVED***,
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *DNSDomainResource***REMOVED*** ImportState(ctx context.Context,
	request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	// To import a dns domain, we need the id (same as the domain itself***REMOVED***
	id := request.ID
	response.Diagnostics.Append(response.State.SetAttribute(ctx, tftypes.NewAttributePath(***REMOVED***.WithAttributeName("id"***REMOVED***, id***REMOVED***...***REMOVED***
}
