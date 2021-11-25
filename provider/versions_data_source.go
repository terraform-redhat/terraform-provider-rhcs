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
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type VersionsDataSourceType struct {
}

type VersionsDataSource struct {
	logger     logging.Logger
	collection *cmv1.VersionsClient
}

func (t *VersionsDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of OpenShift versions.",
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

func (t *VersionsDataSourceType***REMOVED*** itemAttributes(***REMOVED*** map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Description: "Unique identifier of the version. This is what should be " +
				"used when referencing the versions from other places, for " +
				"example in the 'version' attribute of the cluster resource.",
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"name": {
			Description: "Short name of the version, for example '4.1.0'.",
			Type:        types.StringType,
			Computed:    true,
***REMOVED***,
	}
}

func (t *VersionsDataSourceType***REMOVED*** NewDataSource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.DataSource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of versions:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***

	// Create the resource:
	result = &VersionsDataSource{
		logger:     parent.logger,
		collection: collection,
	}
	return
}

func (s *VersionsDataSource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse***REMOVED*** {
	// Get the state:
	state := &VersionsState{}
	diags := request.Config.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Fetch the list of versions:
	var listItems []*cmv1.Version
	listSize := 100
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.Size(listSize***REMOVED***
	if !state.Search.Unknown && !state.Search.Null {
		listRequest.Search(state.Search.Value***REMOVED***
	} else {
		listRequest.Search("enabled = 't'"***REMOVED***
	}
	if !state.Order.Unknown && !state.Order.Null {
		listRequest.Order(state.Order.Value***REMOVED***
	}
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
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
			ID: types.String{
				Value: listItem.ID(***REMOVED***,
	***REMOVED***,
			Name: types.String{
				Value: listItem.RawID(***REMOVED***,
	***REMOVED***,
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
