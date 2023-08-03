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
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const DataSourceID = "groups"

func GroupsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: groupsDataSourceRead,
		Schema:      groupsSchema(),
	}
}

func groupsDataSourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Get the collection of AWSInquiries:
	collection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()

	groupsState := groupsStateFromResourceData(resourceData)
	// Fetch the complete list of groups of the cluster:
	var listItems []*cmv1.Group
	listSize := 10
	listPage := 1
	listRequest := collection.Cluster(groupsState.Cluster).Groups().List()
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Can't list groups",
					Detail:   err.Error(),
				}}
		}
		if listItems == nil {
			listItems = make([]*cmv1.Group, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.Group) bool {
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
	groupsState.Items = make([]GroupState, len(listItems))
	for i, listItem := range listItems {
		groupsState.Items[i] = GroupState{
			ID:   listItem.ID(),
			Name: listItem.ID(),
		}
	}

	resourceData.Set("items", flatGroups(groupsState))

	resourceData.SetId(DataSourceID)
	return
}
