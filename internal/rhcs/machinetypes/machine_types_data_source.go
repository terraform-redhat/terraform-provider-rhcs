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

package machinetypes

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const DataSourceID = "machine-types"

func MachineTypesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: machineTypesDataSourceRead,
		Schema:      machineTypesSchema(),
	}
}

func machineTypesDataSourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Get the collection of machines:
	collection := meta.(*sdk.Connection).ClustersMgmt().V1().MachineTypes()

	var listItems []*cmv1.MachineType
	listSize := 10
	listPage := 1
	listRequest := collection.List().Size(listSize)
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Can't list machine types",
					Detail:   err.Error(),
				}}
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

	listOfMachineTypes := []interface{}{}
	for _, listItem := range listItems {
		machineTypeMap := map[string]interface{}{}

		cpuObject := listItem.CPU()
		cpuValue := cpuObject.Value()
		cpuUnit := cpuObject.Unit()

		switch cpuUnit {
		case "vCPU":
			// Nothing.
		default:
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Unknown CPU unit",
					Detail:   fmt.Sprintf("Don't know how to convert CPU unit '%s'", cpuUnit),
				}}
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
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Unknown RAM unit",
					Detail:   fmt.Sprintf("Don't know how to convert RAM unit '%s'", ramUnit),
				}}
		}
		machineTypeMap["id"] = listItem.ID()
		machineTypeMap["name"] = listItem.Name()
		machineTypeMap["cpu"] = int64(cpuValue)
		machineTypeMap["cloud_provider"] = listItem.CloudProvider().ID()
		machineTypeMap["ram"] = int64(ramValue)

		listOfMachineTypes = append(listOfMachineTypes, machineTypeMap)
	}

	resourceData.SetId(DataSourceID)
	resourceData.Set("items", listOfMachineTypes)

	return
}
