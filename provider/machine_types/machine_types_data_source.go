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

package machine_types

***REMOVED***
	"context"
***REMOVED***
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type MachineTypesDataSource struct {
	collection *cmv1.MachineTypesClient
}

var _ datasource.DataSource = &MachineTypesDataSource{}
var _ datasource.DataSourceWithConfigure = &MachineTypesDataSource{}

func New(***REMOVED*** datasource.DataSource {
	return &MachineTypesDataSource{}
}

func (s *MachineTypesDataSource***REMOVED*** Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_machine_types"
}

func (s *MachineTypesDataSource***REMOVED*** Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse***REMOVED*** {
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
				***REMOVED***,
						"id": schema.StringAttribute{
							Description: "Unique identifier of the machine type.",
							Computed:    true,
				***REMOVED***,
						"name": schema.StringAttribute{
							Description: "Short name of the machine type.",
							Computed:    true,
				***REMOVED***,
						"cpu": schema.Int64Attribute{
							Description: "Number of vCPU cores.",
							Computed:    true,
				***REMOVED***,
						"ram": schema.Int64Attribute{
							Description: "Amount of RAM in bytes.",
							Computed:    true,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
}

func (s *MachineTypesDataSource***REMOVED*** Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection***REMOVED***

	// Get the collection of cloud providers:
	s.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.MachineTypes(***REMOVED***
}

func (s *MachineTypesDataSource***REMOVED*** Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse***REMOVED*** {
	// Fetch the complete list of machine types:
	var listItems []*cmv1.MachineType
	listSize := 10
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.Size(listSize***REMOVED***
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't list machine types",
				err.Error(***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		if listItems == nil {
			listItems = make([]*cmv1.MachineType, 0, listResponse.Total(***REMOVED******REMOVED***
***REMOVED***
		listResponse.Items(***REMOVED***.Each(func(listItem *cmv1.MachineType***REMOVED*** bool {
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
	state := &MachineTypesState{
		Items: make([]*MachineTypeState, len(listItems***REMOVED******REMOVED***,
	}
	for i, listItem := range listItems {
		cpuObject := listItem.CPU(***REMOVED***
		cpuValue := cpuObject.Value(***REMOVED***
		cpuUnit := cpuObject.Unit(***REMOVED***
		switch cpuUnit {
		case "vCPU":
			// Nothing.
		default:
			resp.Diagnostics.AddError(
				"Unknown CPU unit",
				fmt.Sprintf("Don't know how to convert CPU unit '%s'", cpuUnit***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		ramObject := listItem.Memory(***REMOVED***
		ramValue := ramObject.Value(***REMOVED***
		ramUnit := ramObject.Unit(***REMOVED***
		switch strings.ToLower(ramUnit***REMOVED*** {
		case "b":
			// Nothing.
		case "kb":
			ramValue *= math.Pow10(3***REMOVED***
		case "mb":
			ramValue *= math.Pow10(6***REMOVED***
		case "gb":
			ramValue *= math.Pow10(9***REMOVED***
		case "tb":
			ramValue *= math.Pow10(12***REMOVED***
		case "pb":
			ramValue *= math.Pow10(15***REMOVED***
		case "kib":
			ramValue *= math.Pow(2, 10***REMOVED***
		case "mib":
			ramValue *= math.Pow(2, 20***REMOVED***
		case "gib":
			ramValue *= math.Pow(2, 30***REMOVED***
		case "tib":
			ramValue *= math.Pow(2, 40***REMOVED***
		case "pib":
			ramValue *= math.Pow(2, 50***REMOVED***
		default:
			resp.Diagnostics.AddError(
				"Unknown RAM unit",
				fmt.Sprintf("Don't know how to convert RAM unit '%s'", ramUnit***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		state.Items[i] = &MachineTypeState{
			CloudProvider: listItem.CloudProvider(***REMOVED***.ID(***REMOVED***,
			ID:            listItem.ID(***REMOVED***,
			Name:          listItem.Name(***REMOVED***,
			CPU:           int64(cpuValue***REMOVED***,
			RAM:           int64(ramValue***REMOVED***,
***REMOVED***
	}

	// Save the state:
	diags := resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}
