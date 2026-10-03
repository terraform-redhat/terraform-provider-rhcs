// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func testAuthState(t *testing.T) ExternalAuthProviderState {
	t.Helper()
	audiences, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"console"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	return ExternalAuthProviderState{
		Cluster:                   types.StringValue("cluster-1"),
		Name:                      types.StringValue("example"),
		IssuerURL:                 types.StringValue("https://issuer.example.com"),
		IssuerAudiences:           audiences,
		IssuerCA:                  types.StringNull(),
		ClaimMappingGroupsClaim:   types.StringValue("groups"),
		ClaimMappingUsernameClaim: types.StringValue("email"),
		ClaimValidationRule:       types.ListNull(types.StringType),
		ConsoleClientID:           types.StringValue("console"),
		ConsoleClientSecret:       types.StringValue("fixture-placeholder"),
		ExtraScopes:               types.SetNull(types.StringType),
		ID:                        types.StringValue("example"),
	}
}

func jsonAuth(t *testing.T, value *cmv1.ExternalAuth) map[string]any {
	t.Helper()
	var buffer bytes.Buffer
	if err := cmv1.MarshalExternalAuth(value, &buffer); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func testRemoteAuth(t *testing.T, clients ...*cmv1.ExternalAuthClientConfigBuilder) *cmv1.ExternalAuth {
	t.Helper()
	value, err := cmv1.NewExternalAuth().ID("example").Issuer(
		cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console"),
	).Clients(clients...).Build()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestExternalAuthListStringsUsesCommonConversionWithoutSilencingUnknowns(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name      string
		value     types.List
		want      []string
		wantError bool
	}{
		{"known", types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("console"), types.StringValue("other"),
		}), []string{"console", "other"}, false},
		{"empty", types.ListValueMust(types.StringType, []attr.Value{}), []string{}, false},
		{"null", types.ListNull(types.StringType), nil, true},
		{"unknown", types.ListUnknown(types.StringType), nil, true},
		{"unknown element", types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown()}), nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := listStrings(ctx, tc.value)
			if (err != nil) != tc.wantError || (!tc.wantError && !reflect.DeepEqual(got, tc.want)) {
				t.Fatalf("listStrings() = %v, %v; want %v, error=%t", got, err, tc.want, tc.wantError)
			}
		})
	}
}

func TestExternalAuthSetNullUnknownPayloads(t *testing.T) {
	ctx := context.Background()
	for _, value := range []types.Set{types.SetNull(types.StringType), types.SetUnknown(types.StringType)} {
		plan := testAuthState(t)
		plan.IssuerAudiences = value
		if _, err := buildCreate(ctx, plan); err == nil {
			t.Fatalf("required issuer_audiences %v must fail before payload build", value)
		}
	}
	plan := testAuthState(t)
	plan.ClaimValidationRule = types.ListUnknown(types.StringType)
	if _, err := buildCreate(ctx, plan); err == nil {
		t.Fatal("unknown claim_validation_rule must fail before payload build")
	}
	prior := testAuthState(t)
	prior.ClaimValidationRule = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:console")})
	plan = prior
	plan.ClaimValidationRule = types.ListNull(types.StringType)
	patch, changed, err := buildPatch(ctx, plan, prior, testRemoteAuth(t))
	if err != nil || !changed {
		t.Fatalf("clear rules patch: changed=%t, err=%v", changed, err)
	}
	claim := jsonAuth(t, patch)["claim"].(map[string]any)
	rules, ok := claim["validation_rules"].([]any)
	if !ok || len(rules) != 0 {
		t.Fatalf("null claim_validation_rule must explicitly clear validation_rules: %v", claim)
	}
}

func TestExternalAuthCreatePayload(t *testing.T) {
	plan := testAuthState(t)
	value, err := buildCreate(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	got := jsonAuth(t, value)
	issuer := got["issuer"].(map[string]any)
	if issuer["url"] != plan.IssuerURL.ValueString() || issuer["ca"] != nil {
		t.Fatalf("unexpected issuer: %v", issuer)
	}
	client := got["clients"].([]any)[0].(map[string]any)
	if client["id"] != "console" || client["secret"] != "fixture-placeholder" || client["type"] != nil {
		t.Fatalf("unexpected console client: %v", client)
	}
	component := client["component"].(map[string]any)
	if component["name"] != "console" || component["namespace"] != "openshift-console" {
		t.Fatalf("unexpected component: %v", component)
	}
}

func TestExternalAuthIssuerAudienceRefreshIgnoresOrder(t *testing.T) {
	ctx := context.Background()
	state := testAuthState(t)
	state.IssuerAudiences = types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("console"), types.StringValue("other"),
	})
	prior := state
	remote, err := cmv1.NewExternalAuth().ID("example").Issuer(
		cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("other", "console"),
	).Clients(cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console"))).Build()
	if err != nil {
		t.Fatal(err)
	}
	if diags := populateFromAPI(ctx, &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	if !state.IssuerAudiences.Equal(prior.IssuerAudiences) {
		t.Fatalf("reordered audiences caused drift: before=%v after=%v", prior.IssuerAudiences, state.IssuerAudiences)
	}
	if _, changed, err := buildPatch(ctx, state, prior, remote); err != nil || changed {
		t.Fatalf("reordered audiences caused PATCH: changed=%t, err=%v", changed, err)
	}
}

func TestExternalAuthReadPreservesEmptyCollections(t *testing.T) {
	for _, tc := range []struct {
		name string
		json string
	}{
		{
			name: "explicit empty",
			json: `{"id":"example","issuer":{"url":"https://issuer.example.com","audiences":["console"]},` +
				`"claim":{"validation_rules":[]},"clients":[{"id":"console",` +
				`"component":{"name":"console","namespace":"openshift-console"},"extra_scopes":[]}]}`,
		},
		{
			name: "omitted",
			json: `{"id":"example","issuer":{"url":"https://issuer.example.com","audiences":["console"]},` +
				`"clients":[{"id":"console","component":{"name":"console","namespace":"openshift-console"}}]}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			remote, err := cmv1.UnmarshalExternalAuth(tc.json)
			if err != nil {
				t.Fatal(err)
			}
			emptyRules := types.ListValueMust(types.StringType, []attr.Value{})
			emptyScopes := types.SetValueMust(types.StringType, []attr.Value{})
			state := testAuthState(t)
			state.ClaimValidationRule = emptyRules
			state.ExtraScopes = emptyScopes
			if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
				t.Fatal(diags)
			}
			if !state.ClaimValidationRule.Equal(emptyRules) || !state.ExtraScopes.Equal(emptyScopes) {
				t.Fatalf("refresh changed empty collections to null: rules=%v scopes=%v",
					state.ClaimValidationRule, state.ExtraScopes)
			}

			state = testAuthState(t)
			if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
				t.Fatal(diags)
			}
			if !state.ClaimValidationRule.IsNull() || !state.ExtraScopes.IsNull() {
				t.Fatalf("refresh changed null collections to empty: rules=%v scopes=%v",
					state.ClaimValidationRule, state.ExtraScopes)
			}
		})
	}
}

func TestExternalAuthClaimPatchPreservesRemotePrefixes(t *testing.T) {
	for _, tc := range []struct {
		name         string
		groupPrefix  string
		userPrefix   string
		prefixPolicy string
	}{
		{"non-empty", "team-", "user-", "Prefix"},
		{"present empty", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			prior := testAuthState(t)
			plan := prior
			plan.ClaimMappingGroupsClaim = types.StringValue("roles")
			remote, err := cmv1.NewExternalAuth().ID("example").Issuer(
				cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console"),
			).Claim(cmv1.NewExternalAuthClaim().Mappings(cmv1.NewTokenClaimMappings().
				Groups(cmv1.NewGroupsClaim().Claim("groups").Prefix(tc.groupPrefix)).
				UserName(cmv1.NewUsernameClaim().Claim("email").Prefix(tc.userPrefix).
					PrefixPolicy(tc.prefixPolicy)))).Build()
			if err != nil {
				t.Fatal(err)
			}
			patch, changed, err := buildPatch(ctx, plan, prior, remote)
			if err != nil || !changed {
				t.Fatalf("claim PATCH: changed=%t err=%v", changed, err)
			}
			payload := jsonAuth(t, patch)
			claimPayload, ok := payload["claim"].(map[string]any)
			if !ok {
				t.Fatalf("claim PATCH missing claim: %v", payload)
			}
			mappings, ok := claimPayload["mappings"].(map[string]any)
			if !ok {
				t.Fatalf("claim PATCH missing mappings: %v", payload)
			}
			groups, ok := mappings["groups"].(map[string]any)
			if !ok {
				t.Fatalf("claim PATCH missing groups mapping: %v", mappings)
			}
			username, ok := mappings["username"].(map[string]any)
			if !ok {
				t.Fatalf("claim PATCH missing username mapping: %v", mappings)
			}
			if groups["claim"] != "roles" || groups["prefix"] != tc.groupPrefix ||
				username["prefix"] != tc.userPrefix || username["prefix_policy"] != tc.prefixPolicy {
				t.Fatalf("claim PATCH lost remote prefix fields: %v", mappings)
			}
			if _, ok := groups["prefix"]; !ok {
				t.Fatal("groups prefix presence was lost")
			}
			if _, ok := username["prefix"]; !ok {
				t.Fatal("username prefix presence was lost")
			}
			if _, ok := username["prefix_policy"]; !ok {
				t.Fatal("username prefix policy presence was lost")
			}
		})
	}
}

func TestExternalAuthClaimRulePreservesColonsInRequiredValue(t *testing.T) {
	ctx := context.Background()
	plan := testAuthState(t)
	var diags diag.Diagnostics
	plan.ClaimValidationRule, diags = types.ListValueFrom(ctx, types.StringType, []string{"aud:urn:example:scope"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	value, err := buildCreate(ctx, plan)
	if err != nil {
		t.Fatal(err)
	}
	rules := jsonAuth(t, value)["claim"].(map[string]any)["validation_rules"].([]any)
	rule := rules[0].(map[string]any)
	if rule["claim"] != "aud" || rule["required_value"] != "urn:example:scope" {
		t.Fatalf("OCM rule lost the required value suffix: %v", rule)
	}
	state := plan
	if readDiags := populateFromAPI(ctx, &state, value); readDiags.HasError() {
		t.Fatal(readDiags)
	}
	var refreshed []string
	if readDiags := state.ClaimValidationRule.ElementsAs(ctx, &refreshed, false); readDiags.HasError() {
		t.Fatal(readDiags)
	}
	if len(refreshed) != 1 || refreshed[0] != "aud:urn:example:scope" {
		t.Fatalf("refresh lost the full rule: %v", refreshed)
	}
}

func TestExternalAuthClaimBuilderRejectsResolvedMalformedRule(t *testing.T) {
	ctx := context.Background()
	for _, value := range []string{"claim", ":value", "claim:", " \t:value", "claim: \t"} {
		t.Run(value, func(t *testing.T) {
			plan := testAuthState(t)
			var diags diag.Diagnostics
			plan.ClaimValidationRule, diags = types.ListValueFrom(ctx, types.StringType, []string{value})
			if diags.HasError() {
				t.Fatal(diags)
			}
			if _, err := buildCreate(ctx, plan); err == nil {
				t.Fatalf("builder accepted malformed resolved rule %q", value)
			}
		})
	}
}

func TestExternalAuthPatchClearsAndProtectsClients(t *testing.T) {
	ctx := context.Background()
	prior := testAuthState(t)
	prior.IssuerCA = types.StringValue("old CA content")
	prior.ExtraScopes, _ = types.SetValueFrom(ctx, types.StringType, []string{"profile"})
	plan := prior
	plan.IssuerCA = types.StringNull()
	plan.ExtraScopes, _ = types.SetValueFrom(ctx, types.StringType, []string{})
	plan.ConsoleClientSecret = types.StringNull()
	remote := testRemoteAuth(t, cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")))
	patch, changed, err := buildPatch(ctx, plan, prior, remote)
	if err != nil || !changed {
		t.Fatalf("buildPatch: changed=%t, err=%v", changed, err)
	}
	got := jsonAuth(t, patch)
	if got["claim"] != nil || got["id"] != nil {
		t.Fatalf("PATCH should omit unchanged fields: %v", got)
	}
	issuer := got["issuer"].(map[string]any)
	if issuer["ca"] != "" {
		t.Fatalf("CA clear was not explicit: %v", issuer)
	}
	client := got["clients"].([]any)[0].(map[string]any)
	if client["extra_scopes"] == nil || client["secret"] != "" || client["type"] != nil {
		t.Fatalf("client clear was not explicit: %v", client)
	}
	if len(client["extra_scopes"].([]any)) != 0 {
		t.Fatalf("scopes were not cleared: %v", client)
	}
	other := cmv1.NewExternalAuthClientConfig().ID("other").
		Component(cmv1.NewClientComponent().Name("other").Namespace("openshift-console"))
	_, _, err = buildPatch(ctx, plan, prior, testRemoteAuth(t, other))
	if err == nil || !strings.Contains(err.Error(), "other clients") {
		t.Fatalf("expected fail-closed for unmanaged client, got %v", err)
	}
}

func TestExternalAuthPatchImportedConfidentialClient(t *testing.T) {
	prior := testAuthState(t)
	prior.ConsoleClientSecret = types.StringNull()
	plan := prior
	plan.ConsoleClientID = types.StringValue("new-console")
	remote := testRemoteAuth(t, cmv1.NewExternalAuthClientConfig().ID("console").
		Type(cmv1.ExternalAuthClientTypeConfidential).
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")))
	_, _, err := buildPatch(context.Background(), plan, prior, remote)
	if err == nil || !strings.Contains(err.Error(), "after import") {
		t.Fatalf("expected missing imported secret diagnostic, got %v", err)
	}
}

func TestExternalAuthReadKeepsSecretAndFindsConsole(t *testing.T) {
	state := testAuthState(t)
	other := cmv1.NewExternalAuthClientConfig().ID("other").
		Component(cmv1.NewClientComponent().Name("other").Namespace("openshift-console"))
	console := cmv1.NewExternalAuthClientConfig().ID("console").Secret("********").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console"))
	remote := testRemoteAuth(t, other, console)
	diags := populateFromAPI(context.Background(), &state, remote)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if state.ConsoleClientID.ValueString() != "console" || state.ConsoleClientSecret.ValueString() != "fixture-placeholder" {
		t.Fatalf("console mapping lost state secret or selected wrong client: %+v", state)
	}
	remote = testRemoteAuth(t, console, console)
	diags = populateFromAPI(context.Background(), &state, remote)
	if !diags.HasError() {
		t.Fatal("duplicate console clients should be ambiguous")
	}
	state = testAuthState(t)
	changedConsole := cmv1.NewExternalAuthClientConfig().ID("new-console").Secret("********").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console"))
	diags = populateFromAPI(context.Background(), &state, testRemoteAuth(t, changedConsole))
	if diags.HasError() || !state.ConsoleClientSecret.IsNull() {
		t.Fatalf("old client secret must not follow new client ID: state=%+v diagnostics=%v", state, diags)
	}
}

func TestExternalAuthReadRejectsMissingRemoteID(t *testing.T) {
	state := testAuthState(t)
	remote, err := cmv1.NewExternalAuth().Issuer(
		cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console"),
	).Build()
	if err != nil {
		t.Fatal(err)
	}
	diags := populateFromAPI(context.Background(), &state, remote)
	if !diags.HasError() || state.Name.ValueString() != "example" {
		t.Fatalf("missing ID must not overwrite state: state=%+v diagnostics=%v", state, diags)
	}
}
