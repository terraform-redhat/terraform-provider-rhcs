/*
func flatGroups(groupsState *GroupsState) []interface{} {
	listOfGroups := []interface{}{}
	for _, group := range groupsState.Items {
		groupMap := make(map[string]string)
		groupMap["id"] = group.ID
		groupMap["name"] = group.Name
		listOfGroups = append(listOfGroups, groupMap)
	}

	return listOfGroups
}

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
)

type VersionsState struct {
	// optional
	Search *string `tfsdk:"search"`
	Order  *string `tfsdk:"order"`

	// computed
	Item  *VersionState  `tfsdk:"item"`
	Items []VersionState `tfsdk:"items"`
}

type VersionState struct {
	ID   string `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

func versionsDataSourcSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"search": {
			Description: "Search criteria.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"order": {
			Description: "Order criteria.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"item": {
			Description: "Content of the list when there is exactly one item.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: itemSchema(),
			},
			Computed: true,
		},
		"items": {
			Description: "Content of the list.",
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
		"id": {
			Description: "Unique identifier of the version. This is what should be " +
				"used when referencing the versions from other places, for " +
				"example in the 'version' attribute of the cluster resource.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Description: "Short name of the version, for example '4.1.0'.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func versionsStateToResourceData(versionsState *VersionsState, resourceData *schema.ResourceData) {
	if versionsState.Item != nil {
		resourceData.Set("item", []interface{}{flatVersion(*versionsState.Item)})
	}
	resourceData.Set("items", flatVersions(versionsState))
}

func flatVersions(VersionsState *VersionsState) []interface{} {
	if len(VersionsState.Items) == 0 {
		return nil
	}
	result := []interface{}{}

	for _, Version := range VersionsState.Items {
		result = append(result, flatVersion(Version))
	}

	return result
}

func flatVersion(item VersionState) map[string]string {
	result := make(map[string]string)

	result["id"] = item.ID
	result["name"] = item.Name

	return result
}

func versionsStateFromResourceData(resourceData *schema.ResourceData) *VersionsState {
	return &VersionsState{
		Search: common.GetOptionalString(resourceData, "search"),
		Order:  common.GetOptionalString(resourceData, "order"),
	}
}
