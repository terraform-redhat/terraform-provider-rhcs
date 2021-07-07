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

package ocm

***REMOVED***
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

func dataSourceCloudProviders(***REMOVED*** *schema.Resource {
	return &schema.Resource{
		Description: "List of cloud providers.",
		Schema: map[string]*schema.Schema{
			idsKey: {
				Description: "Set of identifiers of the cloud providers.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
		ReadContext: dataSourceCloudProvidersRead,
	}
}

func dataSourceCloudProvidersRead(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Fetch the complete list of cloud providers:
	var providers []*cmv1.CloudProvider
	resource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.CloudProviders(***REMOVED***
	size := 10
	page := 1
	request := resource.List(***REMOVED***.Size(size***REMOVED***
	for {
		response, err := request.SendContext(ctx***REMOVED***
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "can't fetch cloud providers",
				Detail:   err.Error(***REMOVED***,
	***REMOVED******REMOVED***
			return
***REMOVED***
		if providers == nil {
			providers = make([]*cmv1.CloudProvider, 0, response.Total(***REMOVED******REMOVED***
***REMOVED***
		response.Items(***REMOVED***.Each(func(provider *cmv1.CloudProvider***REMOVED*** bool {
			providers = append(providers, provider***REMOVED***
			return true
***REMOVED******REMOVED***
		if response.Size(***REMOVED*** < size {
			break
***REMOVED***
		page++
		request.Page(page***REMOVED***
	}

	// Compute the set of identifiers:
	set := map[string]bool{}
	for _, provider := range providers {
		set[provider.ID(***REMOVED***] = true
	}
	ids := make([]string, 0, len(set***REMOVED******REMOVED***
	for id := range set {
		ids = append(ids, id***REMOVED***
	}
	sort.Strings(ids***REMOVED***

	// Populate the data:
	data.SetId("-"***REMOVED***
	data.Set(idsKey, ids***REMOVED***

	return
}
