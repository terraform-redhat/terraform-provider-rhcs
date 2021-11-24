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
			"items": {
				Description: "Items of the list.",
				Attributes: tfsdk.ListNestedAttributes(
					map[string]tfsdk.Attribute{
						"id": {
							Description: "Unique identifier of the " +
								"version. This is what " +
								"should be used when referencing " +
								"the versions from other " +
								"places, for example in the " +
								"'version' attribute " +
								"of the cluster resource.",
							Type:     types.StringType,
							Computed: true,
				***REMOVED***,
						"name": {
							Description: "Short name of the version " +
								"provider, for example '4.1.0'.",
							Type:     types.StringType,
							Computed: true,
				***REMOVED***,
			***REMOVED***,
					tfsdk.ListNestedAttributesOptions{},
				***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
	return
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
	// Fetch the list of verisions:
	var listItems []*cmv1.Version
	listSize := 100
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.
		Search("enabled = 't'"***REMOVED***.
		Size(listSize***REMOVED***
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
	state := &VersionsState{
		Items: make([]*VersionState, len(listItems***REMOVED******REMOVED***,
	}
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

	// Save the state:
	diags := response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}
