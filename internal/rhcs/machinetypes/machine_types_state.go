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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type MachineTypesState struct {
	Items []MachineTypeState `tfsdk:"items"`
}

type MachineTypeState struct {
	CloudProvider string `tfsdk:"cloud_provider"`
	ID            string `tfsdk:"id"`
	Name          string `tfsdk:"name"`
	CPU           int64  `tfsdk:"cpu"`
	RAM           int64  `tfsdk:"ram"`
}

func machineTypesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"items": {
			Description: "Items of the list.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: itemSchema(),
			},
			Computed: true,
		},
	}

}

func itemSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cloud_provider": {
			Description: "Unique identifier of the " +
				"cloud clusterservice where the machine " +
				"type is supported.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"id": {
			Description: "Unique identifier of the " +
				"machine type.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Description: "Short name of the machine " +
				"type.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"cpu": {
			Description: "Number of CPU cores.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"ram": {
			Description: "Amount of RAM in bytes.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}
}
