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
***REMOVED***
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type MachineTypesDataSourceType struct {
}

type MachineTypesDataSource struct {
	logger     logging.Logger
	collection *cmv1.MachineTypesClient
}

func (t *MachineTypesDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of cloud providers.",
		Attributes: map[string]tfsdk.Attribute{
			"items": {
				Description: "Items of the list.",
				Attributes: tfsdk.ListNestedAttributes(
					map[string]tfsdk.Attribute{
						"cloud_provider": {
							Description: "Unique identifier of the " +
								"cloud provider where the machine " +
								"type is supported.",
							Type:     types.StringType,
							Computed: true,
				***REMOVED***,
						"id": {
							Description: "Unique identifier of the " +
								"machine type.",
							Type:     types.StringType,
							Computed: true,
				***REMOVED***,
						"name": {
							Description: "Short name of the machine " +
								"type.",
							Type:     types.StringType,
							Computed: true,
				***REMOVED***,
						"cpu": {
							Description: "Number of CPU cores.",
							Type:        types.Int64Type,
							Computed:    true,
				***REMOVED***,
						"ram": {
							Description: "Amount of RAM in bytes.",
							Type:        types.Int64Type,
							Computed:    true,
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

func (t *MachineTypesDataSourceType***REMOVED*** NewDataSource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.DataSource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of machine types:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.MachineTypes(***REMOVED***

	// Create the resource:
	result = &MachineTypesDataSource{
		logger:     parent.logger,
		collection: collection,
	}
	return
}

func (s *MachineTypesDataSource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse***REMOVED*** {
	// Fetch the complete list of machine types:
	var listItems []*cmv1.MachineType
	listSize := 10
	listPage := 1
	listRequest := s.collection.List(***REMOVED***.Size(listSize***REMOVED***
	for {
		listResponse, err := listRequest.SendContext(ctx***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
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
			response.Diagnostics.AddError(
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
			response.Diagnostics.AddError(
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
	diags := response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}
