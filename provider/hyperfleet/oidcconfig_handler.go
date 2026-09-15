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

package hyperfleet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
)

// OidcConfigHandlerImpl is the concrete implementation of OidcConfigHandler
type OidcConfigHandlerImpl struct {
	accountID string
	callerARN string
}

// NewOidcConfigHandler creates a new OidcConfigHandler
func NewOidcConfigHandler(accountID, callerARN string) OidcConfigHandler {
	return &OidcConfigHandlerImpl{
		accountID: accountID,
		callerARN: callerARN,
	}
}

// PreExpand validates inputs before expansion
func (h *OidcConfigHandlerImpl) PreExpand(ctx context.Context, input *OidcConfigState) diag.Diagnostics {
	var diags diag.Diagnostics

	// Validate required fields
	if input.Name.IsNull() || input.Name.ValueString() == "" {
		diags.AddError("name is required", "OidcConfig name must be specified")
		return diags
	}

	if input.IssuerUrl.IsNull() || input.IssuerUrl.ValueString() == "" {
		diags.AddError("issuer_url is required", "OIDC issuer URL must be specified")
		return diags
	}

	if input.InstallerRoleArn.IsNull() || input.InstallerRoleArn.ValueString() == "" {
		diags.AddError("installer_role_arn is required", "Installer role ARN must be specified")
		return diags
	}

	if input.SecretArn.IsNull() || input.SecretArn.ValueString() == "" {
		diags.AddError("secret_arn is required", "Secret ARN must be specified")
		return diags
	}

	return diags
}

// PostExpand is called after pathbind.Expand to set unmappable fields
func (h *OidcConfigHandlerImpl) PostExpand(
	ctx context.Context,
	input *OidcConfigState,
	obj *v1alpha1.OidcConfig,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// All OidcConfig fields are already handled by pathbind.Expand
	// No additional SDK-specific setup needed at this time

	return diags
}

// PostResponse is called after API operations
func (h *OidcConfigHandlerImpl) PostResponse(ctx context.Context, resp *v1alpha1.OidcConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	// No post-response processing needed for OidcConfig
	return diags
}

// PostFlatten populates computed fields in the state from the API response
func (h *OidcConfigHandlerImpl) PostFlatten(ctx context.Context, state *OidcConfigState, resp *v1alpha1.OidcConfig) {
	// TODO: Populate computed fields from API response
}

// NewOidcConfigHandlerImpl creates a new OidcConfigHandlerImpl instance.
func NewOidcConfigHandlerImpl(client hyperfleet.Interface, accountID string, callerARN string) OidcConfigHandler {
	return &OidcConfigHandlerImpl{
		accountID: accountID,
		callerARN: callerARN,
	}
}
