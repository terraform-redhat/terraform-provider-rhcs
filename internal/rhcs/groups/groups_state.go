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

package groups

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type GroupsState struct {
	//required
	Cluster string `tfsdk:"cluster"`

	// computed
	Items []GroupState `tfsdk:"items"`
}

type GroupState struct {
	ID   string `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

func groupsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster": {
			Description: "Identifier of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
		},
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
		"id": {
			Description: "Unique identifier of the group. This is what " +
				"should be used when referencing the group from other " +
				"places, for example in the 'group' attribute of the " +
				"user resource.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Description: "Short name of the group for example " +
				"'dedicated-admins'.",
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func groupsStateFromResourceData(resourceData *schema.ResourceData) *GroupsState {
	return &GroupsState{
		Cluster: resourceData.Get("cluster").(string),
		Items:   []GroupState{},
	}
}

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
