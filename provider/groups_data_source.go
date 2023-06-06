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

type GroupsDataSourceType struct {
}

type GroupsDataSource struct {
	collection *cmv1.ClustersClient
}

func (t *GroupsDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of groups.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"items": {
				Description: "Items of the list.",
				Attributes:  t.itemSchema(***REMOVED***,
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *GroupsDataSourceType***REMOVED*** itemSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.ListNestedAttributes(
		map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier of the group. This is what " +
					"should be used when referencing the group from other " +
					"places, for example in the 'group' attribute of the " +
					"user resource.",
				Type:     types.StringType,
				Computed: true,
	***REMOVED***,
			"name": {
				Description: "Short name of the group for example " +
					"'dedicated-admins'.",
				Type:     types.StringType,
				Computed: true,
	***REMOVED***,
***REMOVED***,
		tfsdk.ListNestedAttributesOptions{},
	***REMOVED***
}

func (t *GroupsDataSourceType***REMOVED*** NewDataSource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.DataSource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &GroupsDataSource{
		collection: collection,
	}
	return
}

func (s *GroupsDataSource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse***REMOVED*** {
	// Get the state:
	state := &GroupsState{}
	diags := request.Config.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Fetch the complete list of groups of the cluster:
	var listItems []*cmv1.Group
	listSize := 10
	listPage := 1
	listRequest := s.collection.Cluster(state.Cluster.Value***REMOVED***.Groups(***REMOVED***.List(***REMOVED***
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't list groups",
				err.Error(***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		if listItems == nil {
			listItems = make([]*cmv1.Group, 0, listResponse.Total(***REMOVED******REMOVED***
***REMOVED***
		listResponse.Items(***REMOVED***.Each(func(listItem *cmv1.Group***REMOVED*** bool {
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
	state.Items = make([]*GroupState, len(listItems***REMOVED******REMOVED***
	for i, listItem := range listItems {
		state.Items[i] = &GroupState{
			ID: types.String{
				Value: listItem.ID(***REMOVED***,
	***REMOVED***,
			Name: types.String{
				Value: listItem.ID(***REMOVED***,
	***REMOVED***,
***REMOVED***
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}
