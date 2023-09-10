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

package cloudprovider

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfdschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type CloudProvidersDataSource struct {
	collection *cmv1.CloudProvidersClient
}

var _ datasource.DataSource = &CloudProvidersDataSource{}
var _ datasource.DataSourceWithConfigure = &CloudProvidersDataSource{}

func NewCloudProvidersDataSource(***REMOVED*** datasource.DataSource {
	return &CloudProvidersDataSource{}
}

func (s *CloudProvidersDataSource***REMOVED*** Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_cloud_providers"
}

func (s *CloudProvidersDataSource***REMOVED*** Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse***REMOVED*** {
	resp.Schema = tfdschema.Schema{
		Description: "List of cloud providers.",
		Attributes: map[string]tfdschema.Attribute{
			"search": tfdschema.StringAttribute{
				Description: "Search criteria.",
				Optional:    true,
	***REMOVED***,
			"order": tfdschema.StringAttribute{
				Description: "Order criteria.",
				Optional:    true,
	***REMOVED***,
			"item": tfdschema.SingleNestedAttribute{
				Description: "Content of the list when there is exactly one item.",
				Attributes:  s.itemAttributes(***REMOVED***,
				Computed:    true,
	***REMOVED***,
			"items": tfdschema.ListNestedAttribute{
				Description: "Content of the list.",
				NestedObject: tfdschema.NestedAttributeObject{
					Attributes: s.itemAttributes(***REMOVED***,
		***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (s *CloudProvidersDataSource***REMOVED*** itemAttributes(***REMOVED*** map[string]tfdschema.Attribute {
	return map[string]tfdschema.Attribute{
		"id": tfdschema.StringAttribute{
			Description: "Unique identifier of the cloud provider. This is what " +
				"should be used when referencing the cloud provider from other " +
				"places, for example in the 'cloud_provider' attribute " +
				"of the cluster resource.",
			Computed: true,
***REMOVED***,
		"name": tfdschema.StringAttribute{
			Description: "Short name of the cloud provider, for example 'aws' " +
				"or 'gcp'.",
			Computed: true,
***REMOVED***,
		"display_name": tfdschema.StringAttribute{
			Description: "Human friendly name of the cloud provider, for example " +
				"'AWS' or 'GCP'",
			Computed: true,
***REMOVED***,
	}
}

func (s *CloudProvidersDataSource***REMOVED*** Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection***REMOVED***

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***
}

func (s *CloudProvidersDataSource***REMOVED*** Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse***REMOVED*** {
	// Get the state:
	state := &CloudProvidersState{}
	diags := req.Config.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Fetch the complete list of cloud providers:
	var listItems []*cmv1.CloudProvider
	listSize := 100
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.Size(listSize***REMOVED***
	if !state.Search.IsUnknown(***REMOVED*** && !state.Search.IsNull(***REMOVED*** {
		listRequest.Search(state.Search.ValueString(***REMOVED******REMOVED***
	}
	if !state.Order.IsUnknown(***REMOVED*** && !state.Order.IsNull(***REMOVED*** {
		listRequest.Order(state.Order.ValueString(***REMOVED******REMOVED***
	}
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			resp.Diagnostics.AddError(
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
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}
