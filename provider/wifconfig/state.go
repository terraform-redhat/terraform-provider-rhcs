/*
Copyright (c) 2025 Red Hat, Inc.

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

package wifconfig

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WifConfigState holds the Terraform state for a WIF config.
type WifConfigState struct {
	ID               types.String `tfsdk:"id"`
	DisplayName      types.String `tfsdk:"display_name"`
	Organization     types.String `tfsdk:"organization"`
	OpenshiftVersion types.String `tfsdk:"openshift_version"`

	GCP *WifGcpState `tfsdk:"gcp"`
}

// WifGcpState holds the GCP-specific WIF configuration.
type WifGcpState struct {
	ProjectID              types.String `tfsdk:"project_id"`
	ProjectNumber          types.String `tfsdk:"project_number"`
	RolePrefix             types.String `tfsdk:"role_prefix"`
	FederatedProjectID     types.String `tfsdk:"federated_project_id"`
	FederatedProjectNumber types.String `tfsdk:"federated_project_number"`
	ImpersonatorEmail      types.String `tfsdk:"impersonator_email"`
	WorkloadIdentityPool   types.Object `tfsdk:"workload_identity_pool"`
	ServiceAccounts        types.List   `tfsdk:"service_accounts"`
	Support                types.Object `tfsdk:"support"`
}

// WifPoolState holds the workload identity pool configuration from OCM.
type WifPoolState struct {
	PoolId           types.String              `tfsdk:"pool_id"`
	IdentityProvider *WifIdentityProviderState `tfsdk:"identity_provider"`
}

// WifIdentityProviderState holds the OIDC identity provider configuration from OCM.
type WifIdentityProviderState struct {
	IdentityProviderId types.String `tfsdk:"identity_provider_id"`
	IssuerUrl          types.String `tfsdk:"issuer_url"`
	Jwks               types.String `tfsdk:"jwks"`
	AllowedAudiences   types.List   `tfsdk:"allowed_audiences"`
}

// WifServiceAccountState holds a service account definition from the OCM blueprint.
type WifServiceAccountState struct {
	ServiceAccountId  types.String               `tfsdk:"service_account_id"`
	AccessMethod      types.String               `tfsdk:"access_method"`
	OsdRole           types.String               `tfsdk:"osd_role"`
	Roles             types.List                 `tfsdk:"roles"`
	CredentialRequest *WifCredentialRequestState `tfsdk:"credential_request"`
}

// WifRoleState holds a role definition from the OCM blueprint.
type WifRoleState struct {
	RoleId           types.String `tfsdk:"role_id"`
	Predefined       types.Bool   `tfsdk:"predefined"`
	Permissions      types.List   `tfsdk:"permissions"`
	ResourceBindings types.List   `tfsdk:"resource_bindings"`
}

// WifResourceBindingState holds a resource binding for a role.
type WifResourceBindingState struct {
	Type types.String `tfsdk:"type"`
	Name types.String `tfsdk:"name"`
}

// WifCredentialRequestState holds the credential request (OpenShift namespace and SA names).
type WifCredentialRequestState struct {
	Namespace           types.String `tfsdk:"namespace"`
	ServiceAccountNames types.List   `tfsdk:"service_account_names"`
}

// WifSupportState holds the support access configuration from OCM.
type WifSupportState struct {
	Principal types.String `tfsdk:"principal"`
	Roles     types.List   `tfsdk:"roles"`
}
