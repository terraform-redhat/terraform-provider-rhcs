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

package oidcconfiginput

import "github.com/hashicorp/terraform-plugin-framework/types"

type RosaOidcConfigInputState struct {
	Region               types.String `tfsdk:"region"`
	BucketName           types.String `tfsdk:"bucket_name"`
	DiscoveryDoc         types.String `tfsdk:"discovery_doc"`
	Jwks                 types.String `tfsdk:"jwks"`
	PrivateKey           types.String `tfsdk:"private_key"`
	PrivateKeyFileName   types.String `tfsdk:"private_key_file_name"`
	PrivateKeySecretName types.String `tfsdk:"private_key_secret_name"`
	IssuerUrl            types.String `tfsdk:"issuer_url"`
}
