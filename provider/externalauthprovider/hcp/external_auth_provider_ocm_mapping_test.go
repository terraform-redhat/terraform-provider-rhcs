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
	if err != nil || changed || patch != nil {
		t.Fatalf("null claim_validation_rule must not trigger a PATCH: changed=%t, err=%v", changed, err)
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
	if _, ok := client["extra_scopes"]; ok {
		t.Fatalf("null console_extra_scopes must be omitted from create payload: %v", client)
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
	).Claim(cmv1.NewExternalAuthClaim().Mappings(cmv1.NewTokenClaimMappings().
		Groups(cmv1.NewGroupsClaim().Claim("groups")).
		UserName(cmv1.NewUsernameClaim().Claim("email")))).
		Clients(cmv1.NewExternalAuthClientConfig().ID("console").
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
	if _, changed, err := buildPatch(ctx, configuredPlan(state, prior), prior, remote); err != nil || changed {
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
			if !state.ClaimValidationRule.Equal(emptyRules) || !state.ExtraScopes.Equal(emptyScopes) {
				t.Fatalf("refresh must resolve omitted collections to empty: rules=%v scopes=%v",
					state.ClaimValidationRule, state.ExtraScopes)
			}
		})
	}
}

func TestExternalAuthReadPopulatesComputedScopes(t *testing.T) {
	state := testAuthState(t)
	remote := testRemoteAuth(t, cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
		ExtraScopes("api-managed"))
	if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	if !state.ExtraScopes.Equal(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("api-managed")})) {
		t.Fatalf("computed scopes did not refresh from the API: %v", state.ExtraScopes)
	}
}

func TestExternalAuthReadPopulatesComputedOptionalFields(t *testing.T) {
	state := testAuthState(t)
	state.IssuerCA = types.StringNull()
	state.ClaimMappingGroupsClaim = types.StringNull()
	state.ClaimMappingUsernameClaim = types.StringNull()
	state.ClaimValidationRule = types.ListNull(types.StringType)
	state.ConsoleClientID = types.StringNull()
	state.ConsoleClientSecret = types.StringNull()
	state.ExtraScopes = types.SetNull(types.StringType)
	remoteClaim := cmv1.NewExternalAuthClaim().
		Mappings(cmv1.NewTokenClaimMappings().
			Groups(cmv1.NewGroupsClaim().Claim("api-groups")).
			UserName(cmv1.NewUsernameClaim().Claim("api-user"))).
		ValidationRules(cmv1.NewTokenClaimValidationRule().Claim("aud").RequiredValue("api-value"))
	remote, err := cmv1.NewExternalAuth().ID("example").
		Issuer(cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console").CA("api-ca")).
		Claim(remoteClaim).
		Clients(cmv1.NewExternalAuthClientConfig().ID("console").
			Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
			Secret("********").ExtraScopes("api-scope")).Build()
	if err != nil {
		t.Fatal(err)
	}
	if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	if state.IssuerCA.ValueString() != "api-ca" || state.ClaimMappingGroupsClaim.ValueString() != "api-groups" ||
		state.ClaimMappingUsernameClaim.ValueString() != "api-user" || state.ConsoleClientID.ValueString() != "console" ||
		!state.ClaimValidationRule.Equal(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:api-value")})) ||
		!state.ExtraScopes.Equal(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("api-scope")})) ||
		!state.ConsoleClientSecret.IsNull() {
		t.Fatalf("readable API values must populate state without recovering the secret: %+v", state)
	}
}

func TestExternalAuthReadRefreshesManagedOptionalFieldsIncludingEmpty(t *testing.T) {
	state := testAuthState(t)
	state.IssuerCA = types.StringValue("previous CA")
	state.ClaimMappingGroupsClaim = types.StringValue("previous-groups")
	state.ClaimMappingUsernameClaim = types.StringValue("previous-user")
	state.ClaimValidationRule = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:previous")})
	state.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("previous-scope")})
	remoteClaim := cmv1.NewExternalAuthClaim().
		Mappings(cmv1.NewTokenClaimMappings().
			Groups(cmv1.NewGroupsClaim().Claim("")).
			UserName(cmv1.NewUsernameClaim().Claim("api-user"))).
		ValidationRules()
	remote, err := cmv1.NewExternalAuth().ID("example").
		Issuer(cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console").CA("")).
		Claim(remoteClaim).
		Clients(cmv1.NewExternalAuthClientConfig().ID("console").
			Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
			ExtraScopes()).Build()
	if err != nil {
		t.Fatal(err)
	}
	if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	if state.IssuerCA.ValueString() != "" || state.ClaimMappingGroupsClaim.ValueString() != "" ||
		state.ClaimMappingUsernameClaim.ValueString() != "api-user" ||
		!state.ClaimValidationRule.Equal(types.ListValueMust(types.StringType, []attr.Value{})) ||
		!state.ExtraScopes.Equal(types.SetValueMust(types.StringType, []attr.Value{})) {
		t.Fatalf("managed fields did not refresh from the API, including empty values: %+v", state)
	}
}

func TestExternalAuthReadRefreshesManagedScopesIncludingEmpty(t *testing.T) {
	for _, tc := range []struct {
		name          string
		remoteScopes  []string
		includeScopes bool
		wantScopes    []string
	}{
		{name: "remote drift", remoteScopes: []string{"api-value"}, includeScopes: true, wantScopes: []string{"api-value"}},
		{name: "remote empty", remoteScopes: []string{}, includeScopes: true, wantScopes: []string{}},
		{name: "remote omitted", wantScopes: []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := testAuthState(t)
			state.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("configured")})
			client := cmv1.NewExternalAuthClientConfig().ID("console").
				Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console"))
			if tc.includeScopes {
				client.ExtraScopes(tc.remoteScopes...)
			}
			if diags := populateFromAPI(context.Background(), &state, testRemoteAuth(t, client)); diags.HasError() {
				t.Fatal(diags)
			}
			want := types.SetValueMust(types.StringType, func() []attr.Value {
				values := make([]attr.Value, 0, len(tc.wantScopes))
				for _, scope := range tc.wantScopes {
					values = append(values, types.StringValue(scope))
				}
				return values
			}())
			if !state.ExtraScopes.Equal(want) {
				t.Fatalf("managed scopes = %v, want %v", state.ExtraScopes, want)
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
	plan.ConsoleClientSecret = types.StringValue("")
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
	if got["issuer"] != nil {
		t.Fatalf("null CA must not be sent when clearing other fields: %v", got["issuer"])
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

func TestExternalAuthPatchDoesNotClearScopesWhenConfigBecomesNull(t *testing.T) {
	ctx := context.Background()
	prior := testAuthState(t)
	prior.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
	plan := prior
	plan.ExtraScopes = types.SetNull(types.StringType)
	remote := testRemoteAuth(t, cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
		ExtraScopes("profile"))
	patch, changed, err := buildPatch(ctx, plan, prior, remote)
	if err != nil || changed || patch != nil {
		t.Fatalf("null scopes should not trigger a PATCH: patch=%v changed=%t err=%v", patch, changed, err)
	}
}

func TestExternalAuthPatchDoesNotTouchOptionalFieldsWhenConfigBecomesNull(t *testing.T) {
	ctx := context.Background()
	prior := testAuthState(t)
	prior.IssuerCA = types.StringValue("old CA")
	prior.ClaimMappingGroupsClaim = types.StringValue("groups")
	prior.ClaimMappingUsernameClaim = types.StringValue("email")
	prior.ClaimValidationRule = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:old")})
	prior.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
	plan := prior
	plan.IssuerCA = types.StringNull()
	plan.ClaimMappingGroupsClaim = types.StringNull()
	plan.ClaimMappingUsernameClaim = types.StringNull()
	plan.ClaimValidationRule = types.ListNull(types.StringType)
	plan.ConsoleClientID = types.StringNull()
	plan.ConsoleClientSecret = types.StringNull()
	plan.ExtraScopes = types.SetNull(types.StringType)
	remote := testRemoteAuth(t, cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
		ExtraScopes("profile"))
	patch, changed, err := buildPatch(ctx, plan, prior, remote)
	if err != nil || changed || patch != nil {
		t.Fatalf("null optional fields should not trigger a PATCH: patch=%v changed=%t err=%v", patch, changed, err)
	}
}

func TestExternalAuthClientPatchPreservesRemoteScopesWhenConfigIsNull(t *testing.T) {
	ctx := context.Background()
	prior := testAuthState(t)
	plan := prior
	plan.ConsoleClientSecret = types.StringValue("replacement-placeholder")
	remote := testRemoteAuth(t, cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
		ExtraScopes("api-managed"))
	patch, changed, err := buildPatch(ctx, plan, prior, remote)
	if err != nil || !changed {
		t.Fatalf("secret PATCH: changed=%t, err=%v", changed, err)
	}
	client := jsonAuth(t, patch)["clients"].([]any)[0].(map[string]any)
	if got := client["extra_scopes"].([]any); !reflect.DeepEqual(got, []any{"api-managed"}) {
		t.Fatalf("updating another client field changed unmanaged scopes: %v", client)
	}
}

func TestExternalAuthPatchPreservesOtherRemoteOptionalFields(t *testing.T) {
	ctx := context.Background()
	prior := testAuthState(t)
	prior.ClaimMappingGroupsClaim = types.StringNull()
	prior.ClaimMappingUsernameClaim = types.StringNull()
	prior.ClaimValidationRule = types.ListNull(types.StringType)
	plan := prior
	plan.IssuerURL = types.StringValue("https://issuer-new.example.com")
	plan.ClaimMappingGroupsClaim = types.StringValue("roles")
	remoteClaim := cmv1.NewExternalAuthClaim().
		Mappings(cmv1.NewTokenClaimMappings().
			Groups(cmv1.NewGroupsClaim().Claim("groups").Prefix("team-")).
			UserName(cmv1.NewUsernameClaim().Claim("email").Prefix("user-").PrefixPolicy("Prefix"))).
		ValidationRules(cmv1.NewTokenClaimValidationRule().Claim("aud").RequiredValue("api-managed"))
	remote, err := cmv1.NewExternalAuth().ID("example").
		Issuer(cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console").CA("api-managed-ca")).
		Claim(remoteClaim).
		Clients(cmv1.NewExternalAuthClientConfig().ID("console").
			Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
			ExtraScopes("api-managed-scope")).Build()
	if err != nil {
		t.Fatal(err)
	}
	patch, changed, err := buildPatch(ctx, plan, prior, remote)
	if err != nil || !changed {
		t.Fatalf("optional field PATCH: changed=%t err=%v", changed, err)
	}
	got := jsonAuth(t, patch)
	issuer := got["issuer"].(map[string]any)
	if issuer["url"] != "https://issuer-new.example.com" || issuer["ca"] != "api-managed-ca" {
		t.Fatalf("null issuer_ca was not preserved when updating issuer_url: %v", issuer)
	}
	claim := got["claim"].(map[string]any)
	mappings := claim["mappings"].(map[string]any)
	groups := mappings["groups"].(map[string]any)
	username := mappings["username"].(map[string]any)
	if groups["claim"] != "roles" || groups["prefix"] != "team-" ||
		username["claim"] != "email" || username["prefix"] != "user-" ||
		username["prefix_policy"] != "Prefix" {
		t.Fatalf("null claim_mapping_username_claim or remote prefixes were not preserved: %v", mappings)
	}
	rules := claim["validation_rules"].([]any)
	if len(rules) != 1 || rules[0].(map[string]any)["required_value"] != "api-managed" {
		t.Fatalf("null claim_validation_rule was not preserved: %v", claim)
	}
	if got["clients"] != nil {
		t.Fatalf("unmanaged client fields should not be included in the PATCH: %v", got["clients"])
	}
}

func TestExternalAuthPatchSendsExplicitEmptyValues(t *testing.T) {
	ctx := context.Background()
	prior := testAuthState(t)
	prior.IssuerCA = types.StringValue("old CA")
	prior.ClaimValidationRule = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:old")})
	prior.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
	plan := prior
	plan.IssuerCA = types.StringValue("")
	plan.ClaimMappingGroupsClaim = types.StringValue("")
	plan.ClaimMappingUsernameClaim = types.StringValue("")
	plan.ClaimValidationRule = types.ListValueMust(types.StringType, []attr.Value{})
	plan.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{})
	plan.ConsoleClientSecret = types.StringValue("")
	remoteClaim := cmv1.NewExternalAuthClaim().Mappings(cmv1.NewTokenClaimMappings().
		Groups(cmv1.NewGroupsClaim().Claim("groups")).
		UserName(cmv1.NewUsernameClaim().Claim("email"))).
		ValidationRules(cmv1.NewTokenClaimValidationRule().Claim("aud").RequiredValue("old"))
	remote, err := cmv1.NewExternalAuth().ID("example").
		Issuer(cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console").CA("old CA")).
		Claim(remoteClaim).
		Clients(cmv1.NewExternalAuthClientConfig().ID("console").
			Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).
			Secret("********").ExtraScopes("profile")).Build()
	if err != nil {
		t.Fatal(err)
	}
	patch, changed, err := buildPatch(ctx, plan, prior, remote)
	if err != nil || !changed {
		t.Fatalf("explicit empty values must trigger a PATCH: changed=%t err=%v", changed, err)
	}
	got := jsonAuth(t, patch)
	issuer := got["issuer"].(map[string]any)
	if issuer["ca"] != "" {
		t.Fatalf("empty issuer_ca was not sent: %v", issuer)
	}
	claim := got["claim"].(map[string]any)
	if !reflect.DeepEqual(claim["validation_rules"], []any{}) {
		t.Fatalf("empty claim_validation_rule was not sent: %v", claim)
	}
	mappings := claim["mappings"].(map[string]any)
	if mappings["groups"].(map[string]any)["claim"] != "" || mappings["username"].(map[string]any)["claim"] != "" {
		t.Fatalf("empty claim mappings were not sent: %v", mappings)
	}
	client := got["clients"].([]any)[0].(map[string]any)
	if !reflect.DeepEqual(client["extra_scopes"], []any{}) || client["secret"] != "" {
		t.Fatalf("empty client values were not sent: %v", client)
	}
}

func TestExternalAuthPatchRemovesConsoleClientOnlyForExplicitEmptyPair(t *testing.T) {
	prior := testAuthState(t)
	plan := prior
	plan.ConsoleClientID = types.StringValue("")
	plan.ConsoleClientSecret = types.StringValue("")
	patch, changed, err := buildPatch(context.Background(), plan, prior, testRemoteAuth(t,
		cmv1.NewExternalAuthClientConfig().ID("console").
			Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console"))))
	if err != nil || !changed {
		t.Fatalf("explicit empty client pair must trigger removal: changed=%t err=%v", changed, err)
	}
	clients := jsonAuth(t, patch)["clients"].([]any)
	if len(clients) != 0 {
		t.Fatalf("expected empty clients list to remove the console client, got %v", clients)
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

func TestExternalAuthPatchAlreadyRemovedConsole(t *testing.T) {
	for _, otherPresent := range []bool{false, true} {
		prior := testAuthState(t)
		prior.ConsoleClientID = types.StringValue("")
		prior.ConsoleClientSecret = types.StringValue("")
		plan := prior
		plan.IssuerURL = types.StringValue("https://issuer-new.example.com")
		var clients []*cmv1.ExternalAuthClientConfigBuilder
		if otherPresent {
			clients = append(clients, cmv1.NewExternalAuthClientConfig().ID("other").
				Component(cmv1.NewClientComponent().Name("other").Namespace("other")))
		}
		remote := testRemoteAuth(t, clients...)
		patch, changed, err := buildPatch(context.Background(), plan, prior, remote)
		if err != nil || !changed {
			t.Fatalf("issuer update failed with otherPresent=%t: changed=%t err=%v", otherPresent, changed, err)
		}
		if _, present := jsonAuth(t, patch)["clients"]; present {
			t.Fatal("issuer-only update must omit clients when the console is already absent")
		}
		patch, changed, err = buildPatch(context.Background(), prior, prior, remote)
		if err != nil || changed || patch != nil {
			t.Fatalf("unchanged removal must be a no-op: changed=%t err=%v", changed, err)
		}
	}
}

func TestExternalAuthImportHydratesConfidentialClientWithoutRecoveringSecret(t *testing.T) {
	state := ExternalAuthProviderState{
		Cluster: types.StringValue("cluster-1"), Name: types.StringValue("example"), ID: types.StringValue("example"),
		IssuerURL: types.StringNull(), IssuerCA: types.StringNull(),
		ClaimMappingGroupsClaim: types.StringNull(), ClaimMappingUsernameClaim: types.StringNull(),
		ClaimValidationRule: types.ListNull(types.StringType), ConsoleClientID: types.StringNull(),
		ConsoleClientSecret: types.StringNull(), ExtraScopes: types.SetNull(types.StringType),
	}
	remote, err := cmv1.NewExternalAuth().ID("example").
		Issuer(cmv1.NewTokenIssuer().URL("https://issuer.example.com").Audiences("console").CA("fixture-ca")).
		Claim(cmv1.NewExternalAuthClaim().Mappings(cmv1.NewTokenClaimMappings().
			Groups(cmv1.NewGroupsClaim().Claim("groups")).UserName(cmv1.NewUsernameClaim().Claim("email"))).
			ValidationRules(cmv1.NewTokenClaimValidationRule().Claim("aud").RequiredValue("urn:example"))).
		Clients(cmv1.NewExternalAuthClientConfig().ID("console").
			Type(cmv1.ExternalAuthClientTypeConfidential).Secret("********").
			Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console")).ExtraScopes("profile"),
			cmv1.NewExternalAuthClientConfig().ID("other").
				Component(cmv1.NewClientComponent().Name("other").Namespace("other"))).Build()
	if err != nil {
		t.Fatal(err)
	}
	if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	plan := testAuthState(t)
	plan.IssuerCA = types.StringValue("fixture-ca")
	plan.ClaimValidationRule = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:urn:example")})
	plan.ConsoleClientSecret = types.StringNull()
	plan.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
	patch, changed, err := buildPatch(context.Background(), plan, state, remote)
	if err != nil || changed || patch != nil {
		t.Fatalf("matching readable import values must not require replacement: changed=%t err=%v", changed, err)
	}
	if !state.ConsoleClientSecret.IsNull() {
		t.Fatal("import must not recover the masked console secret")
	}
	plan.ConsoleClientSecret = types.StringValue("fixture-placeholder")
	_, _, err = buildPatch(context.Background(), plan, state, remote)
	if err == nil || !strings.Contains(err.Error(), "other clients") {
		t.Fatalf("supplying an imported secret must retain the other-client protection: %v", err)
	}
	// Omitting configuration still records readable API values, without recovering the secret.
	state.ConsoleClientID = types.StringNull()
	state.ConsoleClientSecret = types.StringNull()
	state.ExtraScopes = types.SetNull(types.StringType)
	if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	if state.ConsoleClientID.ValueString() != "console" || !state.ConsoleClientSecret.IsNull() ||
		!state.ExtraScopes.Equal(plan.ExtraScopes) {
		t.Fatal("ordinary refresh must populate computed console values without a secret")
	}
}

func TestExternalAuthPatchRemovesRecreatedConsole(t *testing.T) {
	prior := testAuthState(t)
	prior.ConsoleClientID = types.StringValue("")
	prior.ConsoleClientSecret = types.StringValue("")
	console := cmv1.NewExternalAuthClientConfig().ID("console").
		Component(cmv1.NewClientComponent().Name("console").Namespace("openshift-console"))
	patch, changed, err := buildPatch(context.Background(), prior, prior, testRemoteAuth(t, console))
	if err != nil || !changed || len(jsonAuth(t, patch)["clients"].([]any)) != 0 {
		t.Fatalf("recreated console must be removed: changed=%t err=%v", changed, err)
	}
	other := cmv1.NewExternalAuthClientConfig().ID("other").
		Component(cmv1.NewClientComponent().Name("other").Namespace("other"))
	_, _, err = buildPatch(context.Background(), prior, prior, testRemoteAuth(t, console, other))
	if err == nil || !strings.Contains(err.Error(), "other clients") {
		t.Fatalf("removal must protect other clients: %v", err)
	}
}

func TestExternalAuthImportResolvesAbsentComputedOptionals(t *testing.T) {
	state := testAuthState(t)
	state.IssuerURL = types.StringNull()
	state.ConsoleClientID = types.StringNull()
	state.ConsoleClientSecret = types.StringNull()
	remote := testRemoteAuth(t)
	if diags := populateFromAPI(context.Background(), &state, remote); diags.HasError() {
		t.Fatal(diags)
	}
	if !state.IssuerCA.Equal(types.StringValue("")) ||
		!state.ClaimMappingGroupsClaim.Equal(types.StringValue("")) ||
		!state.ClaimMappingUsernameClaim.Equal(types.StringValue("")) ||
		!state.ClaimValidationRule.Equal(types.ListValueMust(types.StringType, []attr.Value{})) ||
		!state.ConsoleClientID.Equal(types.StringValue("")) || !state.ConsoleClientSecret.IsNull() ||
		!state.ExtraScopes.Equal(types.SetValueMust(types.StringType, []attr.Value{})) {
		t.Fatal("absent computed API attributes must resolve to empty values; secret stays null")
	}
}

func TestExternalAuthComputedValuesDoNotSatisfyConfigurationDependencies(t *testing.T) {
	for _, tc := range []struct {
		name      string
		id        types.String
		secret    types.String
		scopes    types.Set
		wantError bool
	}{
		{"omitted client is readable", types.StringNull(), types.StringNull(), types.SetNull(types.StringType), false},
		{"secret still requires configured ID", types.StringNull(), types.StringValue("fixture-placeholder"),
			types.SetNull(types.StringType), true},
		{"ID still requires secret", types.StringValue("console"), types.StringNull(),
			types.SetNull(types.StringType), true},
		{"scopes still require configured ID", types.StringNull(), types.StringNull(),
			types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")}), true},
		{"configured pair", types.StringValue("console"), types.StringValue("fixture-placeholder"),
			types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")}), false},
		{"removal ignores computed scopes", types.StringValue(""), types.StringValue(""),
			types.SetNull(types.StringType), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testAuthState(t)
			config.ConsoleClientID, config.ConsoleClientSecret, config.ExtraScopes = tc.id, tc.secret, tc.scopes
			plan := config
			if config.ConsoleClientID.IsNull() {
				plan.ConsoleClientID = types.StringValue("console")
			}
			if config.ExtraScopes.IsNull() {
				plan.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
			}
			var diags diag.Diagnostics
			validateCrossFields(context.Background(), configuredPlan(plan, config), &diags)
			if diags.HasError() != tc.wantError {
				t.Fatalf("configuration dependency changed: %v", diags)
			}
		})
	}
}

func TestExternalAuthCreateOmitsUnconfiguredUnknownComputedValues(t *testing.T) {
	config := testAuthState(t)
	config.IssuerCA = types.StringNull()
	config.ClaimMappingGroupsClaim = types.StringNull()
	config.ClaimMappingUsernameClaim = types.StringNull()
	config.ConsoleClientID = types.StringNull()
	config.ConsoleClientSecret = types.StringNull()
	plan := config
	plan.IssuerCA = types.StringUnknown()
	plan.ClaimMappingGroupsClaim = types.StringUnknown()
	plan.ClaimMappingUsernameClaim = types.StringUnknown()
	plan.ClaimValidationRule = types.ListUnknown(types.StringType)
	plan.ConsoleClientID = types.StringUnknown()
	plan.ExtraScopes = types.SetUnknown(types.StringType)
	body, err := buildCreate(context.Background(), configuredPlan(plan, config))
	if err != nil {
		t.Fatal(err)
	}
	got := jsonAuth(t, body)
	if _, present := got["clients"]; present {
		t.Fatal("an unconfigured computed client must not be created")
	}
	if _, present := got["issuer"].(map[string]any)["ca"]; present {
		t.Fatal("an unconfigured computed CA must not be sent")
	}
	if _, present := got["claim"].(map[string]any)["validation_rules"]; present {
		t.Fatal("unconfigured computed rules must not be sent")
	}
}
