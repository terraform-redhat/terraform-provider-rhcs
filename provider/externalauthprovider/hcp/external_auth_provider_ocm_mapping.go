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
	ctx context.Context, plan ExternalAuthProviderState, createDefaults bool, remote *cmv1.ExternalAuthClaim,
) (*cmv1.ExternalAuthClaimBuilder, error) {
	var remoteGroups *cmv1.GroupsClaim
	var remoteUsername *cmv1.UsernameClaim
	if remote != nil {
		if mappings, ok := remote.GetMappings(); ok {
			remoteGroups, _ = mappings.GetGroups()
			remoteUsername, _ = mappings.GetUserName()
		}
	}
	groupsClaim := plan.ClaimMappingGroupsClaim
	if groupsClaim.IsUnknown() {
		return nil, fmt.Errorf("claim_mapping_groups_claim must be known before building the request")
	}
	if groupsClaim.IsNull() {
		if remoteGroups != nil {
			if claim, ok := remoteGroups.GetClaim(); ok {
				groupsClaim = types.StringValue(claim)
			}
		} else if createDefaults {
			groupsClaim = types.StringValue("groups")
		}
	}
	usernameClaim := plan.ClaimMappingUsernameClaim
	if usernameClaim.IsUnknown() {
		return nil, fmt.Errorf("claim_mapping_username_claim must be known before building the request")
	}
	if usernameClaim.IsNull() {
		if remoteUsername != nil {
			if claim, ok := remoteUsername.GetClaim(); ok {
				usernameClaim = types.StringValue(claim)
			}
		} else if createDefaults {
			usernameClaim = types.StringValue("email")
		}
	}

	mappings := cmv1.NewTokenClaimMappings()
	mappingsSet := false
	if !groupsClaim.IsNull() || remoteGroups != nil {
		groupsBuilder := cmv1.NewGroupsClaim()
		if !groupsClaim.IsNull() {
			groupsBuilder.Claim(groupsClaim.ValueString())
		}
		if remoteGroups != nil {
			if prefix, present := remoteGroups.GetPrefix(); present {
				groupsBuilder.Prefix(prefix)
			}
		}
		mappings.Groups(groupsBuilder)
		mappingsSet = true
	}
	if !usernameClaim.IsNull() || remoteUsername != nil {
		usernameBuilder := cmv1.NewUsernameClaim()
		if !usernameClaim.IsNull() {
			usernameBuilder.Claim(usernameClaim.ValueString())
		}
		if remoteUsername != nil {
			if prefix, present := remoteUsername.GetPrefix(); present {
				usernameBuilder.Prefix(prefix)
			}
			if policy, present := remoteUsername.GetPrefixPolicy(); present {
				usernameBuilder.PrefixPolicy(policy)
			}
		}
		mappings.UserName(usernameBuilder)
		mappingsSet = true
	}
	claim := cmv1.NewExternalAuthClaim()
	if mappingsSet {
		claim.Mappings(mappings)
	}
	if !plan.ClaimValidationRule.IsNull() {
		values, err := listStrings(ctx, plan.ClaimValidationRule)
		if err != nil {
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
	} else if remote != nil {
		if remoteRules, present := remote.GetValidationRules(); present {
			rules := make([]*cmv1.TokenClaimValidationRuleBuilder, 0, len(remoteRules))
			for _, rule := range remoteRules {
				if rule != nil {
					rules = append(rules, cmv1.NewTokenClaimValidationRule().
						Claim(rule.Claim()).RequiredValue(rule.RequiredValue()))
				}
			}
			claim.ValidationRules(rules...)
		}
	}
	return claim, nil
}

func issuerBuilder(
	ctx context.Context, plan ExternalAuthProviderState, existing *cmv1.TokenIssuer,
) (*cmv1.TokenIssuerBuilder, error) {
	audiences, err := setStrings(ctx, plan.IssuerAudiences)
	if err != nil {
		return nil, err
	}
	issuer := cmv1.NewTokenIssuer().URL(plan.IssuerURL.ValueString()).Audiences(audiences...)
	if plan.IssuerCA.IsUnknown() {
		return nil, fmt.Errorf("issuer_ca must be known before building the request")
	}
	if !plan.IssuerCA.IsNull() {
		issuer.CA(plan.IssuerCA.ValueString())
	} else if existing != nil {
		if ca, present := existing.GetCA(); present {
			issuer.CA(ca)
		}
	}
	return issuer, nil
}

func consoleClientBuilder(
	ctx context.Context, plan ExternalAuthProviderState, secret types.String, existing *cmv1.ExternalAuthClientConfig,
) (
	*cmv1.ExternalAuthClientConfigBuilder, error,
) {
	client := cmv1.NewExternalAuthClientConfig().
		ID(plan.ConsoleClientID.ValueString()).
		Component(cmv1.NewClientComponent().Name(consoleComponentName).Namespace(consoleComponentNamespace))
	// Forward the configured secret; OCM rejects an empty secret for an existing console client.
	client.Secret(secret.ValueString())
	if !plan.ExtraScopes.IsNull() {
		scopes, err := setStrings(ctx, plan.ExtraScopes)
		if err != nil {
			return nil, err
		}
		client.ExtraScopes(scopes...)
	} else if existing != nil {
		// Clients is rebuilt when another console-client field changes, so preserve
		// remote scopes when Terraform leaves console_extra_scopes unmanaged.
		if scopes, ok := existing.GetExtraScopes(); ok {
			client.ExtraScopes(scopes...)
		}
	}
	return client, nil
}

func buildCreate(ctx context.Context, plan ExternalAuthProviderState) (*cmv1.ExternalAuth, error) {
	issuer, err := issuerBuilder(ctx, plan, nil)
	if err != nil {
		return nil, err
	}
	claim, err := claimBuilder(ctx, plan, true, nil)
	if err != nil {
		return nil, err
	}
	builder := cmv1.NewExternalAuth().ID(plan.Name.ValueString()).Issuer(issuer).Claim(claim)
	if !plan.ConsoleClientID.IsNull() && plan.ConsoleClientID.ValueString() != "" {
		client, err := consoleClientBuilder(ctx, plan, plan.ConsoleClientSecret, nil)
		if err != nil {
			return nil, err
		}
		builder.Clients(client)
	}
	return builder.Build()
}

func findConsoleClient(remote *cmv1.ExternalAuth) (*cmv1.ExternalAuthClientConfig, int, int) {
	var console *cmv1.ExternalAuthClientConfig
	unknownCount, consoleCount := 0, 0
	for _, client := range remote.Clients() {
		if client == nil {
			continue
		}
		component, ok := client.GetComponent()
		if ok && component.Name() == consoleComponentName && component.Namespace() == consoleComponentNamespace {
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
	issuerCAChanged := !plan.IssuerCA.IsNull() && !plan.IssuerCA.Equal(prior.IssuerCA)
	if !plan.IssuerURL.Equal(prior.IssuerURL) || !plan.IssuerAudiences.Equal(prior.IssuerAudiences) || issuerCAChanged {
		remoteIssuer, _ := remote.GetIssuer()
		issuer, err := issuerBuilder(ctx, plan, remoteIssuer)
		if err != nil {
			return nil, false, err
		}
		builder.Issuer(issuer)
		changed = true
	}
	groupsClaimChanged := !plan.ClaimMappingGroupsClaim.IsNull() &&
		!plan.ClaimMappingGroupsClaim.Equal(prior.ClaimMappingGroupsClaim)
	usernameClaimChanged := !plan.ClaimMappingUsernameClaim.IsNull() &&
		!plan.ClaimMappingUsernameClaim.Equal(prior.ClaimMappingUsernameClaim)
	validationRulesChanged := !plan.ClaimValidationRule.IsNull() &&
		!plan.ClaimValidationRule.Equal(prior.ClaimValidationRule)
	if groupsClaimChanged || usernameClaimChanged || validationRulesChanged {
		remoteClaim, _ := remote.GetClaim()
		claim, err := claimBuilder(ctx, plan, false, remoteClaim)
		if err != nil {
			return nil, false, err
		}
		builder.Claim(claim)
		changed = true
	}
	secretChanged := !plan.ConsoleClientSecret.IsNull() && !plan.ConsoleClientSecret.Equal(prior.ConsoleClientSecret)
	scopesChanged := !plan.ExtraScopes.IsNull() && !plan.ExtraScopes.Equal(prior.ExtraScopes)
	clearConsoleClient := !plan.ConsoleClientID.IsNull() && plan.ConsoleClientID.ValueString() == "" &&
		!plan.ConsoleClientSecret.IsNull() && plan.ConsoleClientSecret.ValueString() == ""
	consoleClientIDChanged := !plan.ConsoleClientID.IsNull() && !plan.ConsoleClientID.Equal(prior.ConsoleClientID)
	console, unknownCount, consoleCount := findConsoleClient(remote)
	consoleChanged := consoleClientIDChanged || scopesChanged || secretChanged
	if clearConsoleClient {
		// An already absent console needs no client-list replacement, even when other fields change.
		consoleChanged = consoleCount > 0
	}
	if consoleChanged {
		if consoleCount > 1 {
			return nil, false, fmt.Errorf("cannot safely update: OCM returned multiple console clients")
		}
		if unknownCount > 0 {
			return nil, false, fmt.Errorf(
				"cannot safely update the console client while other clients are present; update them outside Terraform first",
			)
		}
		if clearConsoleClient {
			builder.Clients()
		} else if !plan.ConsoleClientID.IsNull() {
			if console != nil && plan.ConsoleClientSecret.IsNull() && prior.ConsoleClientSecret.IsNull() &&
				plan.ConsoleClientID.ValueString() != "" {
				return nil, false, fmt.Errorf(
					"cannot update a confidential console client after import without console_client_secret; configure it first",
				)
			}
			client, err := consoleClientBuilder(ctx, plan, plan.ConsoleClientSecret, console)
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
	ca, _ := issuer.GetCA()
	state.IssuerCA = types.StringValue(ca)
	rules := make([]string, 0)
	remoteClaim, _ := remote.GetClaim()
	var mappings *cmv1.TokenClaimMappings
	if remoteClaim != nil {
		mappings, _ = remoteClaim.GetMappings()
	}
	groupsClaim, usernameClaim := "", ""
	if mappings != nil {
		if groups, ok := mappings.GetGroups(); ok {
			groupsClaim = groups.Claim()
		}
		if username, ok := mappings.GetUserName(); ok {
			usernameClaim = username.Claim()
		}
	}
	state.ClaimMappingGroupsClaim = types.StringValue(groupsClaim)
	state.ClaimMappingUsernameClaim = types.StringValue(usernameClaim)
	if remoteClaim != nil {
		for _, rule := range remoteClaim.ValidationRules() {
			if rule != nil {
				rules = append(rules, rule.Claim()+":"+rule.RequiredValue())
			}
		}
	}
	state.ClaimValidationRule, diags = types.ListValueFrom(ctx, types.StringType, rules)
	if diags.HasError() {
		return diags
	}
	console, _, consoleCount := findConsoleClient(remote)
	if consoleCount > 1 {
		diags.AddError("Ambiguous console client", "OCM returned multiple console clients for the same component.")
		return diags
	}
	if console == nil {
		state.ConsoleClientID = types.StringValue("")
		state.ExtraScopes, diags = types.SetValueFrom(ctx, types.StringType, []string{})
		return diags
	}
	if state.ConsoleClientID.ValueString() != console.ID() {
		state.ConsoleClientSecret = types.StringNull()
	}
	state.ConsoleClientID = types.StringValue(console.ID())
	// OCM may omit or mask secret; never copy a returned secret into state.
	scopes := console.ExtraScopes()
	if scopes == nil {
		scopes = []string{}
	}
	state.ExtraScopes, diags = types.SetValueFrom(ctx, types.StringType, scopes)
	return diags
}

// Computed values come from OCM; only explicitly configured values belong in requests and validation.
func configuredPlan(plan, config ExternalAuthProviderState) ExternalAuthProviderState {
	if config.IssuerCA.IsNull() {
		plan.IssuerCA = types.StringNull()
	}
	if config.ClaimMappingGroupsClaim.IsNull() {
		plan.ClaimMappingGroupsClaim = types.StringNull()
	}
	if config.ClaimMappingUsernameClaim.IsNull() {
		plan.ClaimMappingUsernameClaim = types.StringNull()
	}
	if config.ClaimValidationRule.IsNull() {
		plan.ClaimValidationRule = types.ListNull(types.StringType)
	}
	if config.ConsoleClientID.IsNull() {
		plan.ConsoleClientID = types.StringNull()
	}
	if config.ExtraScopes.IsNull() {
		plan.ExtraScopes = types.SetNull(types.StringType)
	}
	return plan
}
