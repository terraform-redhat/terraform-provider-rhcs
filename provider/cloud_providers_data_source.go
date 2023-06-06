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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type CloudProvidersDataSourceType struct {
}

type CloudProvidersDataSource struct {
	collection *cmv1.CloudProvidersClient
}

func (t *CloudProvidersDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of cloud providers.",
		Attributes: map[string]tfsdk.Attribute{
			"search": {
				Description: "Search criteria.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"order": {
				Description: "Order criteria.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"item": {
				Description: "Content of the list when there is exactly one item.",
				Attributes:  tfsdk.SingleNestedAttributes(t.itemAttributes(***REMOVED******REMOVED***,
				Computed:    true,
	***REMOVED***,
			"items": {
				Description: "Content of the list.",
				Attributes: tfsdk.ListNestedAttributes(
					t.itemAttributes(***REMOVED***,
					tfsdk.ListNestedAttributesOptions{},
				***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *CloudProvidersDataSourceType***REMOVED*** itemAttributes(***REMOVED*** map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Description: "Unique identifier of the cloud provider. This is what " +
				"should be used when referencing the cloud provider from other " +
				"places, for example in the 'cloud_provider' attribute " +
				"of the cluster resource.",
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"name": {
			Description: "Short name of the cloud provider, for example 'aws' " +
				"or 'gcp'.",
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"display_name": {
			Description: "Human friendly name of the cloud provider, for example " +
				"'AWS' or 'GCP'",
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
	}
}

func (t *CloudProvidersDataSourceType***REMOVED*** NewDataSource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.DataSource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***

	// Create the resource:
	result = &CloudProvidersDataSource{
		collection: collection,
	}
	return
}

func (s *CloudProvidersDataSource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse***REMOVED*** {
	// Get the state:
	state := &CloudProvidersState{}
	diags := request.Config.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Fetch the complete list of cloud providers:
	var listItems []*cmv1.CloudProvider
	listSize := 100
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.Size(listSize***REMOVED***
	if !state.Search.Unknown && !state.Search.Null {
		listRequest.Search(state.Search.Value***REMOVED***
	}
	if !state.Order.Unknown && !state.Order.Null {
		listRequest.Order(state.Order.Value***REMOVED***
	}
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't list cloud providers",
				err.Error(***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		if listItems == nil {
			listItems = make([]*cmv1.CloudProvider, 0, listResponse.Total(***REMOVED******REMOVED***
***REMOVED***
		listResponse.Items(***REMOVED***.Each(func(listItem *cmv1.CloudProvider***REMOVED*** bool {
			listItems = append(listItems, listItem***REMOVED***
			return true
***REMOVED******REMOVED***
		if listResponse.Size(***REMOVED*** < listSize {
			break
***REMOVED***
		listPage++
		listRequest.Page(listPage***REMOVED***
	}

	// Populate the state:
	state.Items = make([]*CloudProviderState, len(listItems***REMOVED******REMOVED***
	for i, listItem := range listItems {
		state.Items[i] = &CloudProviderState{
			ID:          listItem.ID(***REMOVED***,
			Name:        listItem.Name(***REMOVED***,
			DisplayName: listItem.DisplayName(***REMOVED***,
***REMOVED***
	}
	if len(state.Items***REMOVED*** == 1 {
		state.Item = state.Items[0]
	} else {
		state.Item = nil
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}
