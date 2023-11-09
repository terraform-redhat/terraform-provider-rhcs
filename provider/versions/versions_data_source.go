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

package versions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type VersionsDataSource struct {
	collection *cmv1.VersionsClient
}

var _ datasource.DataSource = &VersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &VersionsDataSource{}

func New() datasource.DataSource {
	return &VersionsDataSource{}
}

func (s *VersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_versions"
}

func (s *VersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of OpenShift versions.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Description: "Search criteria.",
				Optional:    true,
			},
			"order": schema.StringAttribute{
				Description: "Order criteria.",
				Optional:    true,
			},
			"item": schema.SingleNestedAttribute{
				Description: "Content of the list when there is exactly one item.",
				Attributes:  s.itemAttributes(),
				Computed:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "Content of the list.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: s.itemAttributes(),
				},
				Computed: true,
			},
		},
	}
}

func (s *VersionsDataSource) itemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier of the version. This is what should be " +
				"used when referencing the versions from other places, for " +
				"example in the 'version' attribute of the cluster resource.",
			Computed: true,
		},
		"name": schema.StringAttribute{
			Description: "Short name of the version, for example '4.1.0'.",
			Computed:    true,
		},
	}
}

func (s *VersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt().V1().Versions()
}

func (s *VersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the state:
	state := &VersionsState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the list of versions:
	var listItems []*cmv1.Version
	listSize := 100
	listPage := 1
	listRequest := s.collection.List().Size(listSize)
	if !state.Search.IsUnknown() && !state.Search.IsNull() {
		listRequest.Search(state.Search.ValueString())
	} else {
		listRequest.Search("enabled = 't'")
	}
	if !state.Order.IsUnknown() && !state.Order.IsNull() {
		listRequest.Order(state.Order.ValueString())
	}
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't list versions",
				err.Error(),
			)
			return
		}
		if listItems == nil {
			listItems = make([]*cmv1.Version, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.Version) bool {
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
	state.Items = make([]*VersionState, len(listItems))
	for i, listItem := range listItems {
		state.Items[i] = &VersionState{
			ID:   types.StringValue(listItem.ID()),
			Name: types.StringValue(listItem.RawID()),
		}
	}
	if len(state.Items) == 1 {
		state.Item = state.Items[0]
	} else {
		state.Item = nil
	}

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
