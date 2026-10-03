// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"
	"crypto/x509"
	"net/url"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

func issuerURLValidator() validator.String {
	return attrvalidators.NewStringValidator(
		"issuer URL must be an absolute HTTPS URL with a host",
		func(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
				return
			}
			issuerURL := req.ConfigValue.ValueString()
			parsed, err := url.Parse(issuerURL)
			if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.Opaque != "" || parsed.User != nil {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid issuer URL",
					"The issuer URL must be an absolute HTTPS URL with a host and no user information.",
				)
			}
		},
	)
}

func issuerCAValidator() validator.String {
	return attrvalidators.NewStringValidator(
		"value must contain a PEM-encoded CA certificate",
		func(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
				return
			}
			if !x509.NewCertPool().AppendCertsFromPEM([]byte(req.ConfigValue.ValueString())) {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid issuer CA bundle",
					"issuer_ca must contain at least one PEM-encoded X.509 certificate.",
				)
			}
		},
	)
}

func scopeValidator() validator.String {
	return attrvalidators.NewStringValidator(
		"scope must be non-empty and contain no whitespace",
		func(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
				return
			}
			scope := req.ConfigValue.ValueString()
			if scope == "" || strings.IndexFunc(scope, unicode.IsSpace) >= 0 {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid extra scope",
					"Each extra_scopes value must be a single, non-empty scope without whitespace.",
				)
			}
		},
	)
}

func claimValidationRuleValidator() validator.String {
	return attrvalidators.NewStringValidator(
		"rule must have claim:required_value format; the first colon separates claim from value",
		func(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
			if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
				return
			}
			parts := strings.SplitN(req.ConfigValue.ValueString(), ":", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid claim validation rule",
					"Each rule must follow claim:required_value with both parts non-empty; "+
						"the first colon is the separator and later colons are part of required_value.",
				)
			}
		},
	)
}

func (r *ExternalAuthProviderResource) ValidateConfig(
	ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse,
) {
	var config ExternalAuthProviderState
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateCrossFields(ctx, config, &resp.Diagnostics)
}

func validateCrossFields(ctx context.Context, config ExternalAuthProviderState, diags *diag.Diagnostics) {
	if config.ConsoleClientID.IsUnknown() {
		return
	}
	if config.ConsoleClientID.IsNull() {
		if !config.ExtraScopes.IsNull() && !config.ExtraScopes.IsUnknown() && len(config.ExtraScopes.Elements()) > 0 {
			diags.AddAttributeError(
				path.Root("extra_scopes"),
				"Missing console client ID",
				"console_client_id is required when extra_scopes is configured.",
			)
		}
		return
	}
	if config.IssuerAudiences.IsNull() || config.IssuerAudiences.IsUnknown() {
		return
	}
	for _, audience := range config.IssuerAudiences.Elements() {
		if audience.IsUnknown() {
			return
		}
	}
	var audiences []string
	audienceDiags := config.IssuerAudiences.ElementsAs(ctx, &audiences, false)
	diags.Append(audienceDiags...)
	if audienceDiags.HasError() {
		return
	}
	for _, audience := range audiences {
		if audience == config.ConsoleClientID.ValueString() {
			return
		}
	}
	diags.AddAttributeError(
		path.Root("console_client_id"),
		"Console client ID is not an issuer audience",
		"Add console_client_id to issuer_audiences so the console can use tokens from this issuer.",
	)
}

func validateResolvedPlan(ctx context.Context, plan ExternalAuthProviderState, diags *diag.Diagnostics) {
	validateCrossFields(ctx, plan, diags)
	if !plan.ConsoleClientSecret.IsNull() && !plan.ConsoleClientSecret.IsUnknown() && plan.ConsoleClientID.IsNull() {
		diags.AddAttributeError(
			path.Root("console_client_secret"),
			"Missing console client ID",
			"console_client_id is required when console_client_secret is configured.",
		)
	}
}
