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

package dnsdomain

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/ocm/resource"
)

func ResourceDNSDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSDomainCreate,
		ReadContext:   resourceDNSDomainRead,
		DeleteContext: resourceDNSDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: DNSDomainFields(),
	}
}

func resourceDNSDomainCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	// Get the cluster collection:
	collection := meta.(*sdk.Connection).ClustersMgmt().V1().DNSDomains()
	dnsDomain := resource.NewDNSDomain(collection)
	resp, err := dnsDomain.Create()

	if err != nil {
		return diag.Errorf(fmt.Sprintf("Failed to create DNS Domain. %v", err.Error()))
	}

	if id, ok := resp.Body().GetID(); ok {
		resourceData.SetId(id)
		resourceData.Set("dns_id", id)
	} else {
		return diag.Errorf("Failed to create DNS Domain. Failed to get an ID")
	}

	return
}
func resourceDNSDomainRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	// Get the cluster collection:
	collection := meta.(*sdk.Connection).ClustersMgmt().V1().DNSDomains()

	dnsState := DNSDomainState{
		ID:    resourceData.Id(),
		DnsID: resourceData.Get("dns_id").(string),
	}
	// Find the identity provider
	dnsDomain := resource.NewDNSDomain(collection)
	get, err := dnsDomain.Get(dnsState.ID)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("DNS domain (%s) not found, removing from state",
			dnsState.ID,
		))
		resourceData.SetId("")
		return
	} else if err != nil {
		return diag.Errorf(fmt.Sprintf("Can't find DNS domain with identifier '%s', error: %v", dnsState.ID, err.Error()))
	}

	object := get.Body()

	if id, ok := object.GetID(); ok {
		resourceData.SetId(id)
		resourceData.Set("dns_id", id)
	} else {
		return diag.Errorf("Failed to read DNS Domain. Failed to get an ID")
	}
	return
}

func resourceDNSDomainDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	// Get the cluster collection:
	collection := meta.(*sdk.Connection).ClustersMgmt().V1().DNSDomains()

	dnsState := DNSDomainState{
		ID:    resourceData.Id(),
		DnsID: resourceData.Get("dns_ids").(string),
	}

	dnsDomain := resource.NewDNSDomain(collection)
	err := dnsDomain.Delete(dnsState.ID)

	if err != nil {
		return diag.Errorf(fmt.Sprintf("Can't delete DNS domain with identifier '%s', error: %v", dnsState.ID, err.Error()))
	}

	resourceData.SetId("")
	return
}
