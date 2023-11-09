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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RosaOidcConfigState struct {
	Managed          types.Bool   `tfsdk:"managed"`
	SecretARN        types.String `tfsdk:"secret_arn"`
	IssuerUrl        types.String `tfsdk:"issuer_url"`
	InstallerRoleARN types.String `tfsdk:"installer_role_arn"`
	ID               types.String `tfsdk:"id"`
	Thumbprint       types.String `tfsdk:"thumbprint"`
	OIDCEndpointURL  types.String `tfsdk:"oidc_endpoint_url"`
}
