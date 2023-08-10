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

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
)

type DNSDomainResourceType struct {
}

type DNSDomainResource struct {
	collection *cmv1.DNSDomainsClient
}

func (t *DNSDomainResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation:
	// use it directly when needed.
	parent := p.(*Provider)

	//Create DNS Clients - not related to a cluster.
	collection := parent.connection.ClustersMgmt().V1().DNSDomains()

	// Create the resource
	result = &DNSDomainResource{
		collection: collection,
	}

	return
}

func (t *DNSDomainResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "DNS Domain",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier of the DNS Domain",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}

	return
}

func (r *DNSDomainResource) Read(ctx context.Context,
	request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {

	// Get the current state
	state := &DNSDomainState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the identity provider
	dnsDomain := resource.NewDNSDomain(r.collection)
	get, err := dnsDomain.Get(state.ID.Value)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("DNS domain (%s) not found, removing from state",
			state.ID.Value,
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf(
				"Can't find DNS domain with identifier '%s'",
				state.ID.Value,
			),
			err.Error(),
		)
		return
	}

	object := get.Body()

	if id, ok := object.GetID(); ok {
		state.ID = types.String{
			Value: id,
		}
	}

	// Save the state
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *DNSDomainResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {

	// Get the plan
	plan := &DNSDomainState{}
	diags := request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	dnsDomain := resource.NewDNSDomain(r.collection)
	resp, err := dnsDomain.Create()

	if err != nil {
		response.Diagnostics.AddError("Failed to create DNS Domain.", err.Error())
		return
	}

	if id, ok := resp.Body().GetID(); ok {
		plan.ID = types.String{
			Value: id,
		}
	} else {
		response.Diagnostics.AddError("Failed to create DNS Domain.", "Failed to get an ID")
		return
	}

	// Save the state
	diags = response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
}

func (r *DNSDomainResource) Update(ctx context.Context,
	request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse) {
	// Until we support. return an informative error
	response.Diagnostics.AddError("Can't update DNS domain", "Update is currently not supported.")
}

func (r *DNSDomainResource) Delete(ctx context.Context,
	request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
	// Get the state
	state := &DNSDomainState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	dnsDomain := resource.NewDNSDomain(r.collection)
	err := dnsDomain.Delete(state.ID.Value)

	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("Can't delete DNS domain with identifier '%s'", state.ID.Value),
			err.Error(),
		)
		return
	}

	// Remove the state
	response.State.RemoveResource(ctx)
}

func (r *DNSDomainResource) ImportState(ctx context.Context,
	request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	// To import a dns domain, we need the id (same as the domain itself)
	id := request.ID
	response.Diagnostics.Append(response.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), id)...)
}
