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

package oidcconfigscehma

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type RosaOidcConfigState struct {
	// Required
	Managed bool `tfsdk:"managed"`

	// Optional
	SecretARN        *string `tfsdk:"secret_arn"`
	IssuerUrl        *string `tfsdk:"issuer_url"`
	InstallerRoleARN *string `tfsdk:"installer_role_arn"`

	// Computed
	ID              string `tfsdk:"id"`
	Thumbprint      string `tfsdk:"thumbprint"`
	OIDCEndpointURL string `tfsdk:"oidc_endpoint_url"`
}

func OidcConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"managed": {
			Description: "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted) OIDC Configuration",
			Type:        schema.TypeBool,
			Required:    true,
			ForceNew:    true,
		},
		"secret_arn": {
			Description: "Indicates for unmanaged OIDC config, the secret ARN",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"issuer_url": {
			Description: "The bucket URL",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"installer_role_arn": {
			Description: "STS Role ARN with get secrets permission",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"thumbprint": {
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"oidc_endpoint_url": {
			Description: "OIDC Endpoint URL",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
