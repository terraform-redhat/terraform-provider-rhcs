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

type RosaOidcConfigInputState struct {
	// Required
	Region string `tfsdk:"region"`

	// Computed
	BucketName           string `tfsdk:"bucket_name"`
	DiscoveryDoc         string `tfsdk:"discovery_doc"`
	Jwks                 string `tfsdk:"jwks"`
	PrivateKey           string `tfsdk:"private_key"`
	PrivateKeyFileName   string `tfsdk:"private_key_file_name"`
	PrivateKeySecretName string `tfsdk:"private_key_secret_name"`
	IssuerUrl            string `tfsdk:"issuer_url"`
}

func OidcConfigInputFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": {
			Description: "Unique identifier of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"bucket_name": {
			Description: "The S3 bucket name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"discovery_doc": {
			Description: "The discovery document string file",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"jwks": {
			Description: "Json web key set string file",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"private_key": {
			Description: "RSA private key",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"private_key_file_name": {
			Description: "The private key file name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"private_key_secret_name": {
			Description: "The secret name that store the private key",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"issuer_url": {
			Description: "The issuer URL",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
