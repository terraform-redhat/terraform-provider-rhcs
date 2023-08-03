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
package oidcconfig

import (
	"context"
	"fmt"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/oidcconfig/oidcconfigscehma"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rosaoidcconfig "github.com/openshift/rosa/pkg/helper/oidc_config"
)

func ResourceOidcConfigInput() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOidcConfigInputCreate,
		ReadContext:   resourceOidcConfigInputRead,
		UpdateContext: nil,
		DeleteContext: resourceOidcConfigInputDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: oidcconfigscehma.OidcConfigInputFields(),
	}
}

func resourceOidcConfigInputCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin create()")
	region := resourceData.Get("region").(string)

	oidcConfigInput, err := rosaoidcconfig.BuildOidcConfigInput("", region)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't generate oidc config input object",
				Detail: fmt.Sprintf(
					"Can't generate oidc config input object: %v",
					err,
				),
			}}
	}
	oidcConfigInputToResourceData(ctx, oidcConfigInput, resourceData)
	return nil
}

func resourceOidcConfigInputRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin read()")

	return nil
}

func resourceOidcConfigInputDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin delete()")

	resourceData.SetId("")
	return nil
}

func oidcConfigInputToResourceData(ctx context.Context, oidcConfigInput rosaoidcconfig.OidcConfigInput, resourceData *schema.ResourceData) {
	// The resource ID is the bucket name.
	// Amazon S3 supports global buckets, which means that each bucket name must be unique across all AWS accounts in all the AWS Regions within a partition
	resourceData.SetId(oidcConfigInput.BucketName)
	resourceData.Set("bucket_name", oidcConfigInput.BucketName)
	resourceData.Set("discovery_doc", oidcConfigInput.DiscoveryDocument)
	resourceData.Set("jwks", string(oidcConfigInput.Jwks[:]))
	resourceData.Set("private_key", string(oidcConfigInput.PrivateKey[:]))
	resourceData.Set("private_key_file_name", oidcConfigInput.PrivateKeyFilename)
	resourceData.Set("private_key_secret_name", oidcConfigInput.PrivateKeySecretName)
	resourceData.Set("issuer_url", oidcConfigInput.IssuerUrl)
}
