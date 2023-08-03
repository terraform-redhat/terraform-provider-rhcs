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
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
)

const DataSourceID = "cloud-provider"

func CloudProvidersDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: cloudProvidersDataSourceRead,
		Schema:      cloudProviderSchema(),
	}
}

func cloudProvidersDataSourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Get the cloud providers collection:
	cloudProivderCollection := meta.(*sdk.Connection).ClustersMgmt().V1().CloudProviders()

	cloudProvidersState := cloudProvidersStateFromResourceData(resourceData)

	// Fetch the complete list of cloud providers:
	var listItems []*cmv1.CloudProvider
	listSize := 100
	listPage := 1
	listRequest := cloudProivderCollection.List().Size(listSize)
	if !common.IsStringAttributeEmpty(cloudProvidersState.Search) {
		listRequest.Search(*cloudProvidersState.Search)
	}
	if !common.IsStringAttributeEmpty(cloudProvidersState.Order) {
		listRequest.Order(*cloudProvidersState.Order)
	}

	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Can't list cloud providers",
					Detail:   err.Error(),
				}}
		}
		if listItems == nil {
			listItems = make([]*cmv1.CloudProvider, 0, listResponse.Total())
		}
		listResponse.Items().Each(func(listItem *cmv1.CloudProvider) bool {
			listItems = append(listItems, listItem)
			return true
		})
		if listResponse.Size() < listSize {
			break
		}
		listPage++
		listRequest.Page(listPage)
	}

	cloudProvidersState.Items = make([]CloudProviderState, len(listItems))
	for i, listItem := range listItems {
		cloudProvidersState.Items[i] = CloudProviderState{
			ID:          listItem.ID(),
			Name:        listItem.Name(),
			DisplayName: listItem.DisplayName(),
		}
	}

	if len(cloudProvidersState.Items) == 1 {
		cloudProvidersState.Item = &cloudProvidersState.Items[0]
	} else {
		cloudProvidersState.Item = nil
	}

	resourceData.SetId(DataSourceID)
	cloudProvidersStateToResourceData(cloudProvidersState, resourceData)
	return nil
}
