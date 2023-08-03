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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
)

func cloudProviderSchema() map[string]*schema.Schema {
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
				Schema: itemAttributes(),
			},
			Computed: true,
		},
		"items": {
			Description: "Content of the list.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: itemAttributes(),
			},
			Computed: true,
		},
	}
}
func itemAttributes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: "Unique identifier of the cloud clusterservice. This is what " +
				"should be used when referencing the cloud clusterservice from other " +
				"places, for example in the 'cloud_provider' attribute " +
				"of the cluster resource.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Description: "Short name of the cloud clusterservice, for example 'aws' " +
				"or 'gcp'.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"display_name": {
			Description: "Human friendly name of the cloud clusterservice, for example " +
				"'AWS' or 'GCP'",
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

type CloudProvidersState struct {
	// optional
	Search *string `tfsdk:"search"`
	Order  *string `tfsdk:"order"`

	// computed
	Item  *CloudProviderState  `tfsdk:"item"`
	Items []CloudProviderState `tfsdk:"items"`
}

type CloudProviderState struct {
	ID          string `tfsdk:"id"`
	Name        string `tfsdk:"name"`
	DisplayName string `tfsdk:"display_name"`
}

func cloudProvidersStateToResourceData(cloudProvidersState *CloudProvidersState, resourceData *schema.ResourceData) {
	if cloudProvidersState.Item != nil {
		resourceData.Set("item", []interface{}{flatCloudProvider(*cloudProvidersState.Item)})
	}
	resourceData.Set("items", flatCloudProviders(cloudProvidersState))
}

func flatCloudProviders(cloudProvidersState *CloudProvidersState) []interface{} {
	if len(cloudProvidersState.Items) == 0 {
		return nil
	}
	result := []interface{}{}

	for _, cloudProvider := range cloudProvidersState.Items {
		result = append(result, flatCloudProvider(cloudProvider))
	}

	return result
}

func flatCloudProvider(item CloudProviderState) map[string]string {
	result := make(map[string]string)

	result["id"] = item.ID
	result["name"] = item.Name
	result["display_name"] = item.DisplayName

	return result
}

func cloudProvidersStateFromResourceData(resourceData *schema.ResourceData) *CloudProvidersState {
	return &CloudProvidersState{
		Search: common.GetOptionalString(resourceData, "search"),
		Order:  common.GetOptionalString(resourceData, "order"),
		Items:  ExpandItemsFromResourceData(resourceData),
		Item:   ExpandItemFromResourceData(resourceData),
	}
}

func ExpandItemsFromResourceData(resourceData *schema.ResourceData) []CloudProviderState {
	result := []CloudProviderState{}
	itemInterface, ok := resourceData.GetOk("item")
	if !ok {
		return result
	}

	l, ok := itemInterface.([]interface{})
	if !ok {
		return result
	}
	for _, cloudProvider := range l {
		cloudProviderMap := cloudProvider.(map[string]interface{})
		item := CloudProviderState{}
		if id := common.GetOptionalStringFromMapString(cloudProviderMap, "id"); id != nil {
			item.ID = *id
		}
		if name := common.GetOptionalStringFromMapString(cloudProviderMap, "name"); name != nil {
			item.Name = *name
		}
		if displayName := common.GetOptionalStringFromMapString(cloudProviderMap, "display_name"); displayName != nil {
			item.DisplayName = *displayName
		}
		result = append(result, item)
	}

	return result
}

func ExpandItemFromResourceData(resourceData *schema.ResourceData) *CloudProviderState {
	result := CloudProviderState{}
	itemInterface, ok := resourceData.GetOk("item")
	if !ok {
		return nil
	}

	l := itemInterface.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	item := l[0].(map[string]interface{})

	if id := common.GetOptionalStringFromMapString(item, "id"); id != nil {
		result.ID = *id
	}
	if name := common.GetOptionalStringFromMapString(item, "name"); name != nil {
		result.Name = *name
	}
	if displayName := common.GetOptionalStringFromMapString(item, "display_name"); displayName != nil {
		result.DisplayName = *displayName
	}

	return &result
}
