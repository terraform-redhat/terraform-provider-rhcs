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

package versions

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type VersionsDataSource struct {
	collection *cmv1.VersionsClient
}

var _ datasource.DataSource = &VersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &VersionsDataSource{}

func New(***REMOVED*** datasource.DataSource {
	return &VersionsDataSource{}
}

func (s *VersionsDataSource***REMOVED*** Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_versions"
}

func (s *VersionsDataSource***REMOVED*** Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "List of OpenShift versions.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Description: "Search criteria.",
				Optional:    true,
	***REMOVED***,
			"order": schema.StringAttribute{
				Description: "Order criteria.",
				Optional:    true,
	***REMOVED***,
			"item": schema.SingleNestedAttribute{
				Description: "Content of the list when there is exactly one item.",
				Attributes:  s.itemAttributes(***REMOVED***,
				Computed:    true,
	***REMOVED***,
			"items": schema.ListNestedAttribute{
				Description: "Content of the list.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: s.itemAttributes(***REMOVED***,
		***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
}

func (t *VersionsDataSource***REMOVED*** itemAttributes(***REMOVED*** map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier of the version. This is what should be " +
				"used when referencing the versions from other places, for " +
				"example in the 'version' attribute of the cluster resource.",
			Computed: true,
***REMOVED***,
		"name": schema.StringAttribute{
			Description: "Short name of the version, for example '4.1.0'.",
			Computed:    true,
***REMOVED***,
	}
}

func (s *VersionsDataSource***REMOVED*** Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection***REMOVED***

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***
}

func (s *VersionsDataSource***REMOVED*** Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse***REMOVED*** {
	// Get the state:
	state := &VersionsState{}
	diags := req.Config.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Fetch the list of versions:
	var listItems []*cmv1.Version
	listSize := 100
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.Size(listSize***REMOVED***
	if !state.Search.IsUnknown(***REMOVED*** && !state.Search.IsNull(***REMOVED*** {
		listRequest.Search(state.Search.ValueString(***REMOVED******REMOVED***
	} else {
		listRequest.Search("enabled = 't'"***REMOVED***
	}
	if !state.Order.IsUnknown(***REMOVED*** && !state.Order.IsNull(***REMOVED*** {
		listRequest.Order(state.Order.ValueString(***REMOVED******REMOVED***
	}
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't list versions",
				err.Error(***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		if listItems == nil {
			listItems = make([]*cmv1.Version, 0, listResponse.Total(***REMOVED******REMOVED***
***REMOVED***
		listResponse.Items(***REMOVED***.Each(func(listItem *cmv1.Version***REMOVED*** bool {
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
	state.Items = make([]*VersionState, len(listItems***REMOVED******REMOVED***
	for i, listItem := range listItems {
		state.Items[i] = &VersionState{
			ID:   types.StringValue(listItem.ID(***REMOVED******REMOVED***,
			Name: types.StringValue(listItem.RawID(***REMOVED******REMOVED***,
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
