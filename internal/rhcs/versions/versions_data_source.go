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

package versions

import (
	"context"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const DataSourceID = "versions"

func VersionsDataSourc() *schema.Resource {
	return &schema.Resource{
		ReadContext: versionsDataSourcRead,
		Schema:      versionsDataSourcSchema(),
	}
}

func versionsDataSourcRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Get the collection of machines:
	collection := meta.(*sdk.Connection).ClustersMgmt().V1().Versions()

	versionsState := versionsStateFromResourceData(resourceData)

	// Fetch the list of versions:
	var listItems []*cmv1.Version
	listSize := 100
	listPage := 1
	listRequest := collection.List().Size(listSize)
	if !common.IsStringAttributeEmpty(versionsState.Search) {
		listRequest.Search(*versionsState.Search)
	} else {
		listRequest.Search("enabled = 't'")
	}
	if !common.IsStringAttributeEmpty(versionsState.Order) {
		listRequest.Order(*versionsState.Order)
	}

	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Can't list versions",
					Detail:   err.Error(),
				}}
		}

		if listItems == nil {
			listItems = make([]*cmv1.Version, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.Version) bool {
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
	versionsState.Items = make([]VersionState, len(listItems))
	for i, listItem := range listItems {
		versionsState.Items[i] = VersionState{
			ID:   listItem.ID(),
			Name: listItem.RawID(),
		}
	}
	if len(versionsState.Items) == 1 {
		versionsState.Item = &versionsState.Items[0]
	} else {
		versionsState.Item = nil
	}

	resourceData.SetId(DataSourceID)
	versionsStateToResourceData(versionsState, resourceData)
	return nil
}
