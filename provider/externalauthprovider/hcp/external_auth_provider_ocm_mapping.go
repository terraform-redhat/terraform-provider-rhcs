// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

const (
	consoleComponentName      = "console"
	consoleComponentNamespace = "openshift-console"
)

func listStrings(ctx context.Context, value types.List) ([]string, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, fmt.Errorf("invalid list value: list must be known and non-null")
	}
	stringsValue, err := common.StringListToArray(ctx, value)
	if err != nil {
		return nil, fmt.Errorf("invalid list value: %w", err)
	}
	return stringsValue, nil
}

func setStrings(ctx context.Context, value types.Set) ([]string, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, fmt.Errorf("invalid set value: set must be known and non-null")
	}
	var stringsValue []string
	if diags := value.ElementsAs(ctx, &stringsValue, false); diags.HasError() {
		return nil, fmt.Errorf("invalid set value: %s", diags[0].Detail())
	}
	return stringsValue, nil
}

func claimBuilder(
	ctx context.Context, plan ExternalAuthProviderState, clearRules bool, remote *cmv1.ExternalAuthClaim,
) (*cmv1.ExternalAuthClaimBuilder, error) {
	groupsBuilder := cmv1.NewGroupsClaim().Claim(plan.ClaimMappingGroupsClaim.ValueString())
	usernameBuilder := cmv1.NewUsernameClaim().Claim(plan.ClaimMappingUsernameClaim.ValueString())
	if remote != nil {
		if mappings, ok := remote.GetMappings(); ok {
			if groups, ok := mappings.GetGroups(); ok {
				if prefix, present := groups.GetPrefix(); present {
					groupsBuilder.Prefix(prefix)
				}
			}
			if username, ok := mappings.GetUserName(); ok {
				if prefix, present := username.GetPrefix(); present {
					usernameBuilder.Prefix(prefix)
				}
				if policy, present := username.GetPrefixPolicy(); present {
					usernameBuilder.PrefixPolicy(policy)
				}
			}
		}
	}
	mappings := cmv1.NewTokenClaimMappings().
		Groups(groupsBuilder).
		UserName(usernameBuilder)
	claim := cmv1.NewExternalAuthClaim().Mappings(mappings)
	if !plan.ClaimValidationRule.IsNull() || clearRules {
		values, err := listStrings(ctx, plan.ClaimValidationRule)
		if err != nil && !plan.ClaimValidationRule.IsNull() {
			return nil, err
		}
		rules := make([]*cmv1.TokenClaimValidationRuleBuilder, 0, len(values))
		for _, value := range values {
			parts := strings.SplitN(value, ":", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
				return nil, fmt.Errorf("claim_validation_rule must have non-empty claim:required_value parts")
			}
			rules = append(rules, cmv1.NewTokenClaimValidationRule().Claim(parts[0]).RequiredValue(parts[1]))
		}
		claim.ValidationRules(rules...)
	}
	return claim, nil
}

func issuerBuilder(
	ctx context.Context, plan ExternalAuthProviderState, clearCA bool,
) (*cmv1.TokenIssuerBuilder, error) {
	audiences, err := setStrings(ctx, plan.IssuerAudiences)
	if err != nil {
		return nil, err
	}
	issuer := cmv1.NewTokenIssuer().URL(plan.IssuerURL.ValueString()).Audiences(audiences...)
	if !plan.IssuerCA.IsNull() || clearCA {
		issuer.CA(plan.IssuerCA.ValueString())
	}
	return issuer, nil
}

func consoleClientBuilder(ctx context.Context, plan ExternalAuthProviderState, secret types.String, clearScopes bool) (
	*cmv1.ExternalAuthClientConfigBuilder, error,
) {
	client := cmv1.NewExternalAuthClientConfig().
		ID(plan.ConsoleClientID.ValueString()).
		Component(cmv1.NewClientComponent().Name(consoleComponentName).Namespace(consoleComponentNamespace))
	// ROSA CLI always sends a secret field; an empty value denotes a public client.
	client.Secret(secret.ValueString())
	if !plan.ExtraScopes.IsNull() || clearScopes {
		var scopes []string
		if !plan.ExtraScopes.IsNull() {
			var err error
			scopes, err = setStrings(ctx, plan.ExtraScopes)
			if err != nil {
				return nil, err
			}
		}
		client.ExtraScopes(scopes...)
	}
	return client, nil
}

func buildCreate(ctx context.Context, plan ExternalAuthProviderState) (*cmv1.ExternalAuth, error) {
	issuer, err := issuerBuilder(ctx, plan, false)
	if err != nil {
		return nil, err
	}
	claim, err := claimBuilder(ctx, plan, false, nil)
	if err != nil {
		return nil, err
	}
	builder := cmv1.NewExternalAuth().ID(plan.Name.ValueString()).Issuer(issuer).Claim(claim)
	if !plan.ConsoleClientID.IsNull() {
		client, err := consoleClientBuilder(ctx, plan, plan.ConsoleClientSecret, false)
		if err != nil {
			return nil, err
		}
		builder.Clients(client)
	}
	return builder.Build()
}

func isConsoleClient(client *cmv1.ExternalAuthClientConfig) bool {
	component, ok := client.GetComponent()
	return ok && component.Name() == consoleComponentName && component.Namespace() == consoleComponentNamespace
}

func findConsoleClient(remote *cmv1.ExternalAuth) (*cmv1.ExternalAuthClientConfig, int, int) {
	var console *cmv1.ExternalAuthClientConfig
	unknownCount, consoleCount := 0, 0
	for _, client := range remote.Clients() {
		if client == nil {
			continue
		}
		if isConsoleClient(client) {
			console = client
			consoleCount++
		} else {
			unknownCount++
		}
	}
	return console, unknownCount, consoleCount
}

func buildPatch(
	ctx context.Context, plan, prior ExternalAuthProviderState, remote *cmv1.ExternalAuth,
) (*cmv1.ExternalAuth, bool, error) {
	builder := cmv1.NewExternalAuth()
	changed := false
	if !plan.IssuerURL.Equal(prior.IssuerURL) || !plan.IssuerAudiences.Equal(prior.IssuerAudiences) ||
		!plan.IssuerCA.Equal(prior.IssuerCA) {
		issuer, err := issuerBuilder(ctx, plan, true)
		if err != nil {
			return nil, false, err
		}
		builder.Issuer(issuer)
		changed = true
	}
	if !plan.ClaimMappingGroupsClaim.Equal(prior.ClaimMappingGroupsClaim) ||
		!plan.ClaimMappingUsernameClaim.Equal(prior.ClaimMappingUsernameClaim) ||
		!plan.ClaimValidationRule.Equal(prior.ClaimValidationRule) {
		remoteClaim, _ := remote.GetClaim()
		claim, err := claimBuilder(ctx, plan, true, remoteClaim)
		if err != nil {
			return nil, false, err
		}
		builder.Claim(claim)
		changed = true
	}
	secretChanged := !plan.ConsoleClientSecret.Equal(prior.ConsoleClientSecret)
	if !plan.ConsoleClientID.Equal(prior.ConsoleClientID) || !plan.ExtraScopes.Equal(prior.ExtraScopes) || secretChanged {
		console, unknownCount, consoleCount := findConsoleClient(remote)
		if consoleCount > 1 {
			return nil, false, fmt.Errorf("cannot safely update: OCM returned multiple console clients")
		}
		if unknownCount > 0 {
			return nil, false, fmt.Errorf(
				"cannot safely update the console client while other clients are present; update them outside Terraform first",
			)
		}
		if plan.ConsoleClientID.IsNull() {
			builder.Clients()
		} else {
			if console != nil && plan.ConsoleClientSecret.IsNull() && prior.ConsoleClientSecret.IsNull() &&
				console.Type() != cmv1.ExternalAuthClientTypePublic {
				return nil, false, fmt.Errorf(
					"cannot update a confidential console client after import without console_client_secret; configure it first",
				)
			}
			client, err := consoleClientBuilder(ctx, plan, plan.ConsoleClientSecret, true)
			if err != nil {
				return nil, false, err
			}
			builder.Clients(client)
		}
		changed = true
	}
	if !changed {
		return nil, false, nil
	}
	object, err := builder.Build()
	return object, true, err
}

func populateFromAPI(
	ctx context.Context, state *ExternalAuthProviderState, remote *cmv1.ExternalAuth,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if remote == nil || remote.ID() == "" {
		diags.AddError("Incomplete external authentication provider", "OCM did not return a provider ID.")
		return diags
	}
	issuer, ok := remote.GetIssuer()
	if !ok || issuer == nil || issuer.URL() == "" || len(issuer.Audiences()) == 0 {
		diags.AddError("Incomplete external authentication provider", "OCM did not return a usable issuer URL and audiences.")
		return diags
	}
	state.Name = types.StringValue(remote.ID())
	state.ID = types.StringValue(remote.ID())
	state.IssuerURL = types.StringValue(issuer.URL())
	state.IssuerAudiences, diags = types.SetValueFrom(ctx, types.StringType, issuer.Audiences())
	if diags.HasError() {
		return diags
	}
	if ca, ok := issuer.GetCA(); ok && ca != "" {
		state.IssuerCA = types.StringValue(ca)
	} else {
		state.IssuerCA = types.StringNull()
	}
	state.ClaimMappingGroupsClaim = types.StringValue("groups")
	state.ClaimMappingUsernameClaim = types.StringValue("email")
	rules := make([]string, 0)
	if claim, ok := remote.GetClaim(); ok && claim != nil {
		if mappings, ok := claim.GetMappings(); ok && mappings != nil {
			if groups, ok := mappings.GetGroups(); ok && groups.Claim() != "" {
				state.ClaimMappingGroupsClaim = types.StringValue(groups.Claim())
			}
			if username, ok := mappings.GetUserName(); ok && username.Claim() != "" {
				state.ClaimMappingUsernameClaim = types.StringValue(username.Claim())
			}
		}
		for _, rule := range claim.ValidationRules() {
			if rule != nil {
				rules = append(rules, rule.Claim()+":"+rule.RequiredValue())
			}
		}
	}
	if len(rules) > 0 || (!state.ClaimValidationRule.IsNull() && len(state.ClaimValidationRule.Elements()) == 0) {
		state.ClaimValidationRule, diags = types.ListValueFrom(ctx, types.StringType, rules)
	} else {
		state.ClaimValidationRule = types.ListNull(types.StringType)
	}
	if diags.HasError() {
		return diags
	}
	console, _, consoleCount := findConsoleClient(remote)
	if consoleCount > 1 {
		diags.AddError("Ambiguous console client", "OCM returned multiple console clients for the same component.")
		return diags
	}
	if console == nil {
		state.ConsoleClientID = types.StringNull()
		state.ConsoleClientSecret = types.StringNull()
		state.ExtraScopes = types.SetNull(types.StringType)
		return diags
	}
	if state.ConsoleClientID.ValueString() != console.ID() {
		state.ConsoleClientSecret = types.StringNull()
	}
	state.ConsoleClientID = types.StringValue(console.ID())
	// OCM may omit or mask secret; never copy a returned secret into state.
	scopes := console.ExtraScopes()
	if len(scopes) > 0 || (!state.ExtraScopes.IsNull() && len(state.ExtraScopes.Elements()) == 0) {
		if scopes == nil {
			scopes = []string{}
		}
		state.ExtraScopes, diags = types.SetValueFrom(ctx, types.StringType, scopes)
	} else {
		state.ExtraScopes = types.SetNull(types.StringType)
	}
	return diags
}
