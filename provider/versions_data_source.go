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

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

type VersionsDataSourceType struct {
}

type VersionsDataSource struct {
	logger     logging.Logger
	collection *cmv1.VersionsClient
}

func (t *VersionsDataSourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "List of OpenShift versions.",
		Attributes: map[string]tfsdk.Attribute{
			"items": {
				Description: "Items of the list.",
				Attributes: tfsdk.ListNestedAttributes(
					map[string]tfsdk.Attribute{
						"id": {
							Description: "Unique identifier of the " +
								"version. This is what " +
								"should be used when referencing " +
								"the versions from other " +
								"places, for example in the " +
								"'version' attribute " +
								"of the cluster resource.",
							Type:     types.StringType,
							Computed: true,
						},
						"name": {
							Description: "Short name of the version " +
								"provider, for example '4.1.0'.",
							Type:     types.StringType,
							Computed: true,
						},
					},
					tfsdk.ListNestedAttributesOptions{},
				),
				Computed: true,
			},
		},
	}
	return
}

func (t *VersionsDataSourceType) NewDataSource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.DataSource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider)

	// Get the collection of versions:
	collection := parent.connection.ClustersMgmt().V1().Versions()

	// Create the resource:
	result = &VersionsDataSource{
		logger:     parent.logger,
		collection: collection,
	}
	return
}

func (s *VersionsDataSource) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse) {
	// Fetch the list of verisions:
	var listItems []*cmv1.Version
	listSize := 100
	listPage := 1
	listRequest := s.collection.List().
		Search("enabled = 't'").
		Size(listSize)
	for {
		listResponse, err := listRequest.SendContext(ctx)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't list versions",
				err.Error(),
			)
			return
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
	state := &VersionsState{
		Items: make([]*VersionState, len(listItems)),
	}
	for i, listItem := range listItems {
		state.Items[i] = &VersionState{
			ID: types.String{
				Value: listItem.ID(),
			},
			Name: types.String{
				Value: listItem.RawID(),
			},
		}
	}

	// Save the state:
	diags := response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}
