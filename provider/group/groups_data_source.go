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

package group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type GroupsDataSource struct {
	collection *cmv1.ClustersClient
}

var _ datasource.DataSource = &GroupsDataSource{}
var _ datasource.DataSourceWithConfigure = &GroupsDataSource{}

func New() datasource.DataSource {
	return &GroupsDataSource{}
}

func (g *GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (g *GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of groups.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "Content of the list.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: g.itemAttributes(),
				},
				Computed: true,
			},
		},
	}
	return
}

func (g *GroupsDataSource) itemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier of the group. This is what " +
				"should be used when referencing the group from other " +
				"places, for example in the 'group' attribute of the " +
				"user resource.",
			Computed: true,
		},
		"name": schema.StringAttribute{
			Description: "Short name of the group for example " +
				"'dedicated-admins'.",
			Computed: true,
		},
	}
}

func (g *GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	g.collection = connection.ClustersMgmt().V1().Clusters()
}

func (g *GroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest,
	response *datasource.ReadResponse) {
	// Get the state:
	state := &GroupsState{}
	diags := request.Config.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Fetch the complete list of groups of the cluster:
	var listItems []*cmv1.Group
	listSize := 10
	listPage := 1
	listRequest := g.collection.Cluster(state.Cluster.ValueString()).Groups().List()
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't list groups",
				err.Error(),
			)
			return
		}
		if listItems == nil {
			listItems = make([]*cmv1.Group, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.Group) bool {
			listItems = append(listItems, listItem)
			return true
		})
		if listResponse.Size() < listSize {
			break
		}
		listPage++
		listRequest.Page(listPage)
	}

	// Populate the state:
	state.Items = make([]*GroupState, len(listItems))
	for i, listItem := range listItems {
		state.Items[i] = &GroupState{
			ID:   types.StringValue(listItem.ID()),
			Name: types.StringValue(listItem.ID()),
		}
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}
