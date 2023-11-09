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

package machine_types

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type MachineTypesDataSource struct {
	collection *cmv1.MachineTypesClient
}

var _ datasource.DataSource = &MachineTypesDataSource{}
var _ datasource.DataSourceWithConfigure = &MachineTypesDataSource{}

func New() datasource.DataSource {
	return &MachineTypesDataSource{}
}

func (s *MachineTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_types"
}

func (s *MachineTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of machine types",
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				Description: "Items of the list.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cloud_provider": schema.StringAttribute{
							Description: "Unique identifier of the cloud provider where the machine type is supported.",
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "Unique identifier of the machine type.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Short name of the machine type.",
							Computed:    true,
						},
						"cpu": schema.Int64Attribute{
							Description: "Number of vCPU cores.",
							Computed:    true,
						},
						"ram": schema.Int64Attribute{
							Description: "Amount of RAM in bytes.",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (s *MachineTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt().V1().MachineTypes()
}

func (s *MachineTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Fetch the complete list of machine types:
	var listItems []*cmv1.MachineType
	listSize := 10
	listPage := 1
	listRequest := s.collection.List().Size(listSize)
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't list machine types",
				err.Error(),
			)
			return
		}
		if listItems == nil {
			listItems = make([]*cmv1.MachineType, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.MachineType) bool {
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
	state := &MachineTypesState{
		Items: make([]*MachineTypeState, len(listItems)),
	}
	for i, listItem := range listItems {
		cpuObject := listItem.CPU()
		cpuValue := cpuObject.Value()
		cpuUnit := cpuObject.Unit()
		switch cpuUnit {
		case "vCPU":
			// Nothing.
		default:
			resp.Diagnostics.AddError(
				"Unknown CPU unit",
				fmt.Sprintf("Don't know how to convert CPU unit '%s'", cpuUnit),
			)
			return
		}
		ramObject := listItem.Memory()
		ramValue := ramObject.Value()
		ramUnit := ramObject.Unit()
		switch strings.ToLower(ramUnit) {
		case "b":
			// Nothing.
		case "kb":
			ramValue *= math.Pow10(3)
		case "mb":
			ramValue *= math.Pow10(6)
		case "gb":
			ramValue *= math.Pow10(9)
		case "tb":
			ramValue *= math.Pow10(12)
		case "pb":
			ramValue *= math.Pow10(15)
		case "kib":
			ramValue *= math.Pow(2, 10)
		case "mib":
			ramValue *= math.Pow(2, 20)
		case "gib":
			ramValue *= math.Pow(2, 30)
		case "tib":
			ramValue *= math.Pow(2, 40)
		case "pib":
			ramValue *= math.Pow(2, 50)
		default:
			resp.Diagnostics.AddError(
				"Unknown RAM unit",
				fmt.Sprintf("Don't know how to convert RAM unit '%s'", ramUnit),
			)
			return
		}
		state.Items[i] = &MachineTypeState{
			CloudProvider: listItem.CloudProvider().ID(),
			ID:            listItem.ID(),
			Name:          listItem.Name(),
			CPU:           int64(cpuValue),
			RAM:           int64(ramValue),
		}
	}

	// Save the state:
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
