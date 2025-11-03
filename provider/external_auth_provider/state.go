package external_auth_provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type State struct {
	Cluster types.String       `tfsdk:"cluster"`
	ID      types.String       `tfsdk:"id"`
	Issuer  *TokenIssuer       `tfsdk:"issuer"`
	Clients types.List         `tfsdk:"clients"`
	Claim   *ExternalAuthClaim `tfsdk:"claim"`
}

type TokenIssuer struct {
	URL       types.String `tfsdk:"url"`
	Audiences types.Set    `tfsdk:"audiences"`
	CA        types.String `tfsdk:"ca"`
}

type ExternalAuthClaim struct {
	Mappings        *TokenClaimMappings `tfsdk:"mappings"`
	ValidationRules types.List          `tfsdk:"validation_rules"`
}

type TokenClaimMappings struct {
	Username *UsernameClaim `tfsdk:"username"`
	Groups   *GroupsClaim   `tfsdk:"groups"`
}

type UsernameClaim struct {
	Claim        types.String `tfsdk:"claim"`
	Prefix       types.String `tfsdk:"prefix"`
	PrefixPolicy types.String `tfsdk:"prefix_policy"`
}

type GroupsClaim struct {
	Claim  types.String `tfsdk:"claim"`
	Prefix types.String `tfsdk:"prefix"`
}

type TokenClaimValidationRule struct {
	Claim         types.String `tfsdk:"claim"`
	RequiredValue types.String `tfsdk:"required_value"`
}

type ExternalAuthClientConfig struct {
	Component   *ClientComponent `tfsdk:"component"`
	ID          types.String     `tfsdk:"id"`
	Secret      types.String     `tfsdk:"secret"`
	ExtraScopes types.Set        `tfsdk:"extra_scopes"`
	Type        types.String     `tfsdk:"type"`
}

type ClientComponent struct {
	Name      types.String `tfsdk:"name"`
	Namespace types.String `tfsdk:"namespace"`
}