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

package trusted_ip_addresses

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type TrustedIpsDataSource struct {
	collection *cmv1.TrustedIpsClient
}

var _ datasource.DataSource = &TrustedIpsDataSource{}
var _ datasource.DataSourceWithConfigure = &TrustedIpsDataSource{}

func New() datasource.DataSource {
	return &TrustedIpsDataSource{}
}

func (s *TrustedIpsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trusted_ip_addresses"
}

func (s *TrustedIpsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of trusted IP addresses",
		Attributes: map[string]schema.Attribute{
			"total": schema.Int64Attribute{
				Description: "Total number of items in the result set.",
				Computed:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of all trusted IP addresses.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "IP address.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Indicates if the IP is enabled.",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (s *TrustedIpsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt().V1().TrustedIPAddresses()
}

func (s *TrustedIpsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Fetch the complete list of trusted IP addresses:
	var listItems []*cmv1.TrustedIp
	listSize := 10
	listPage := 1
	listRequest := s.collection.List().Size(listSize)

	// Initialize variables to track total items and current page
	var totalItems int

	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't list trusted IP addresses",
				err.Error(),
			)
			return
		}

		// Update total items and current page values
		totalItems = listResponse.Total()

		if listItems == nil {
			listItems = make([]*cmv1.TrustedIp, 0, totalItems)
		}
		listResponse.Items().Each(func(listItem *cmv1.TrustedIp) bool {
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
	state := &TrustedIpsState{
		Total: basetypes.NewInt64Value(int64(totalItems)),
		Items: make([]*TrustedIpState, 0),
	}

	state.Items = fromOCMListItems(listItems)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func fromOCMListItems(listItems []*cmv1.TrustedIp) []*TrustedIpState {
	var trustedIpStates []*TrustedIpState
	for _, listItem := range listItems {
		trustedIpState := &TrustedIpState{
			ID:      basetypes.NewStringValue(listItem.ID()),
			Enabled: basetypes.NewBoolValue(listItem.Enabled()),
		}
		trustedIpStates = append(trustedIpStates, trustedIpState)
	}
	return trustedIpStates
}
