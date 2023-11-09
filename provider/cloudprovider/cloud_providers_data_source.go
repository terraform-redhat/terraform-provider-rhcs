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

package cloudprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type CloudProvidersDataSource struct {
	collection *cmv1.CloudProvidersClient
}

var _ datasource.DataSource = &CloudProvidersDataSource{}
var _ datasource.DataSourceWithConfigure = &CloudProvidersDataSource{}

func New() datasource.DataSource {
	return &CloudProvidersDataSource{}
}

func (s *CloudProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_providers"
}

func (s *CloudProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of cloud providers.",
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

func (s *CloudProvidersDataSource) itemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier of the cloud provider. This is what " +
				"should be used when referencing the cloud provider from other " +
				"places, for example in the 'cloud_provider' attribute " +
				"of the cluster resource.",
			Computed: true,
		},
		"name": schema.StringAttribute{
			Description: "Short name of the cloud provider, for example 'aws' " +
				"or 'gcp'.",
			Computed: true,
		},
		"display_name": schema.StringAttribute{
			Description: "Human friendly name of the cloud provider, for example " +
				"'AWS' or 'GCP'",
			Computed: true,
		},
	}
}

func (s *CloudProvidersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt().V1().CloudProviders()
}

func (s *CloudProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the state:
	state := &CloudProvidersState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the complete list of cloud providers:
	var listItems []*cmv1.CloudProvider
	listSize := 100
	listPage := 1
	listRequest := s.collection.List().Size(listSize)
	if !state.Search.IsUnknown() && !state.Search.IsNull() {
		listRequest.Search(state.Search.ValueString())
	}
	if !state.Order.IsUnknown() && !state.Order.IsNull() {
		listRequest.Order(state.Order.ValueString())
	}
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't list cloud providers",
				err.Error(),
			)
			return
		}
		if listItems == nil {
			listItems = make([]*cmv1.CloudProvider, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.CloudProvider) bool {
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
	state.Items = make([]*CloudProviderState, len(listItems))
	for i, listItem := range listItems {
		state.Items[i] = &CloudProviderState{
			ID:          listItem.ID(),
			Name:        listItem.Name(),
			DisplayName: listItem.DisplayName(),
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
