// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func TestExternalAuthProviderSchema(t *testing.T) {
	t.Parallel()
	r := New()
	if _, ok := r.(*ExternalAuthProviderResource); !ok {
		t.Fatalf("New() returned %T", r)
	}

	metadata := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "rhcs"}, metadata)
	if metadata.TypeName != "rhcs_external_auth_provider" {
		t.Fatalf("unexpected type name: %q", metadata.TypeName)
	}

	response := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, response)
	if response.Diagnostics.HasError() {
		t.Fatalf("invalid schema: %v", response.Diagnostics)
	}
	attrs := response.Schema.Attributes
	if len(attrs) != 12 {
		t.Fatalf("expected 12 attributes, got %d", len(attrs))
	}
	for _, name := range []string{"cluster", "name", "issuer_url"} {
		if a, ok := attrs[name].(schema.StringAttribute); !ok || !a.Required {
			t.Errorf("%s must be a required string", name)
		}
	}
	for _, name := range []string{"cluster", "name"} {
		if len(attrs[name].(schema.StringAttribute).PlanModifiers) == 0 {
			t.Errorf("%s must require replacement on change", name)
		}
	}
	if a, ok := attrs["issuer_audiences"].(schema.SetAttribute); !ok || !a.Required || a.ElementType != types.StringType {
		t.Error("issuer_audiences must be a required set of strings")
	}
	if a, ok := attrs["claim_validation_rule"].(schema.ListAttribute); !ok || !a.Optional || a.ElementType != types.StringType {
		t.Error("claim_validation_rule must be an optional list of strings")
	}
	if a, ok := attrs["console_extra_scopes"].(schema.SetAttribute); !ok || !a.Optional || a.ElementType != types.StringType {
		t.Error("console_extra_scopes must be an optional set of strings")
	}
	for _, name := range []string{"issuer_ca", "console_client_id", "console_client_secret"} {
		if a, ok := attrs[name].(schema.StringAttribute); !ok || !a.Optional {
			t.Errorf("%s must be an optional string", name)
		}
	}
	for _, name := range []string{
		"issuer_ca", "claim_mapping_groups_claim", "claim_mapping_username_claim",
		"claim_validation_rule", "console_client_id", "console_extra_scopes",
	} {
		if !attrs[name].IsOptional() || !attrs[name].IsComputed() {
			t.Errorf("%s must be optional and computed", name)
		}
	}
	if attrs["console_client_secret"].IsComputed() {
		t.Error("console_client_secret must remain optional without computed")
	}
	if attrs["issuer_ca"].(schema.StringAttribute).Sensitive {
		t.Error("issuer_ca must not be sensitive")
	}
	if !attrs["console_client_secret"].(schema.StringAttribute).Sensitive {
		t.Error("console_client_secret must be sensitive")
	}
	if len(attrs["console_client_secret"].(schema.StringAttribute).Validators) == 0 {
		t.Error("console_client_secret needs non-empty validation")
	}
	if a, ok := attrs["id"].(schema.StringAttribute); !ok || !a.Computed {
		t.Error("id must be a computed string")
	}
	for _, name := range []string{"claim_mapping_groups_claim", "claim_mapping_username_claim"} {
		a := attrs[name].(schema.StringAttribute)
		if !a.Optional || !a.Computed || a.Default != nil {
			t.Errorf("%s must be optional and computed without a schema default", name)
		}
	}

	stateType := reflect.TypeOf(ExternalAuthProviderState{})
	if stateType.NumField() != len(attrs) {
		t.Fatalf("state has %d fields; schema has %d attributes", stateType.NumField(), len(attrs))
	}
	for i := 0; i < stateType.NumField(); i++ {
		field := stateType.Field(i)
		name := field.Tag.Get("tfsdk")
		attribute, ok := attrs[name]
		if !ok {
			t.Errorf("state field %s has no matching schema attribute %q", field.Name, name)
			continue
		}
		switch attribute.(type) {
		case schema.ListAttribute:
			if field.Type != reflect.TypeOf(types.List{}) {
				t.Errorf("%s needs types.List", name)
			}
		case schema.SetAttribute:
			if field.Type != reflect.TypeOf(types.Set{}) {
				t.Errorf("%s needs types.Set", name)
			}
		default:
			if field.Type != reflect.TypeOf(types.String{}) {
				t.Errorf("%s needs types.String", name)
			}
		}
	}
}

func TestExternalAuthProviderStringValidators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		validator validator.String
		value     string
		valid     bool
	}{
		{"https issuer", issuerURLValidator(), "https://issuer.example.com/realm", true},
		{"http issuer", issuerURLValidator(), "http://issuer.example.com", false},
		{"issuer without host", issuerURLValidator(), "https:///realm", false},
		{"issuer without scheme", issuerURLValidator(), "issuer.example.com", false},
		{"issuer with credentials", issuerURLValidator(), "https://user:pass@issuer.example.com", false},
		{"valid claim rule", claimValidationRuleValidator(), "email:required@example.com", true},
		{"value with colons", claimValidationRuleValidator(), "aud:urn:example:scope", true},
		{"extra colon in value", claimValidationRuleValidator(), "claim:value:extra", true},
		{"empty claim", claimValidationRuleValidator(), ":value", false},
		{"whitespace claim", claimValidationRuleValidator(), " \t:value", false},
		{"empty required value", claimValidationRuleValidator(), "claim:", false},
		{"whitespace required value", claimValidationRuleValidator(), "claim: \t ", false},
		{"no separator", claimValidationRuleValidator(), "claim", false},
		{"single scope", scopeValidator(), "profile", true},
		{"empty scope", scopeValidator(), "", false},
		{"scope with space", scopeValidator(), "profile email", false},
		{"scope with tab", scopeValidator(), "profile\temail", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			response := &validator.StringResponse{}
			tc.validator.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("field"),
				ConfigValue: types.StringValue(tc.value),
			}, response)
			if response.Diagnostics.HasError() == tc.valid {
				t.Errorf("value %q valid=%t, diagnostics=%v", tc.value, tc.valid, response.Diagnostics)
			}
		})
	}
}

func TestExternalAuthProviderNonBlankSchemaValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	response := &resource.SchemaResponse{}
	New().Schema(ctx, resource.SchemaRequest{}, response)
	state := tfsdk.State{Schema: response.Schema}
	if diags := state.Set(ctx, testAuthState(t)); diags.HasError() {
		t.Fatalf("set test config: %v", diags)
	}
	config := tfsdk.Config{Schema: response.Schema, Raw: state.Raw}

	for _, name := range []string{
		"cluster",
		"name",
		"claim_mapping_groups_claim",
		"claim_mapping_username_claim",
		"console_client_id",
		"console_client_secret",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			validators := response.Schema.Attributes[name].(schema.StringAttribute).Validators
			if len(validators) == 0 {
				t.Fatal("expected a schema string validator")
			}
			emptyIsClear := name == "claim_mapping_groups_claim" || name == "claim_mapping_username_claim" ||
				name == "console_client_id" || name == "console_client_secret"
			for _, tc := range []struct {
				name  string
				value string
				valid bool
			}{
				{"nonblank", "provider one", true},
				{"empty clear value", "", emptyIsClear},
				{"ASCII whitespace", " \t\r\n", false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					var diags diag.Diagnostics
					for _, v := range validators {
						result := &validator.StringResponse{}
						v.ValidateString(context.Background(), validator.StringRequest{
							Path:           path.Root(name),
							PathExpression: path.MatchRoot(name),
							Config:         config,
							ConfigValue:    types.StringValue(tc.value),
						}, result)
						diags.Append(result.Diagnostics...)
					}
					if diags.HasError() == tc.valid {
						t.Errorf("value %q valid=%t, diagnostics=%v", tc.value, tc.valid, diags)
					}
				})
			}
		})
	}
}

func TestIssuerCAValidation(t *testing.T) {
	t.Parallel()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, publicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	validPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	for _, tc := range []struct {
		name  string
		value string
		valid bool
	}{
		{"valid certificate", validPEM, true},
		{"non PEM text", "not a certificate", false},
		{"empty content clears CA", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			response := &validator.StringResponse{}
			issuerCAValidator().ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("issuer_ca"),
				ConfigValue: types.StringValue(tc.value),
			}, response)
			if response.Diagnostics.HasError() == tc.valid {
				t.Errorf("valid=%t, diagnostics=%v", tc.valid, response.Diagnostics)
			}
		})
	}
}

func TestExternalAuthProviderIssuerAudiencesValidation(t *testing.T) {
	t.Parallel()
	r := New()
	response := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, response)
	validators := response.Schema.Attributes["issuer_audiences"].(schema.SetAttribute).Validators
	for _, tc := range []struct {
		name      string
		audiences []string
		valid     bool
	}{
		{"empty", nil, false},
		{"one", []string{"console"}, true},
		{"ten", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, true},
		{"eleven", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}, false},
		{"empty item", []string{"console", ""}, false},
		{"ASCII whitespace item", []string{"console", " \t\r\n"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			values := make([]attr.Value, 0, len(tc.audiences))
			for _, audience := range tc.audiences {
				values = append(values, types.StringValue(audience))
			}
			audienceSet := types.SetValueMust(types.StringType, values)
			var diags diag.Diagnostics
			for _, v := range validators {
				result := &validator.SetResponse{}
				v.ValidateSet(context.Background(), validator.SetRequest{
					Path: path.Root("issuer_audiences"), ConfigValue: audienceSet,
				}, result)
				diags.Append(result.Diagnostics...)
			}
			if diags.HasError() == tc.valid {
				t.Errorf("audiences %v valid=%t, diagnostics=%v", tc.audiences, tc.valid, diags)
			}
		})
	}
}

func TestExternalAuthProviderOptionalCollectionValidation(t *testing.T) {
	t.Parallel()
	r := New()
	response := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, response)

	rules := response.Schema.Attributes["claim_validation_rule"].(schema.ListAttribute)
	invalidRules := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("claim: \t")})
	var ruleDiags diag.Diagnostics
	for _, v := range rules.Validators {
		result := &validator.ListResponse{}
		v.ValidateList(context.Background(), validator.ListRequest{Path: path.Root("claim_validation_rule"), ConfigValue: invalidRules}, result)
		ruleDiags.Append(result.Diagnostics...)
	}
	if !ruleDiags.HasError() {
		t.Error("malformed claim rule must be rejected by the schema")
	}
	validRules := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("aud:urn:example:scope")})
	ruleDiags = nil
	for _, v := range rules.Validators {
		result := &validator.ListResponse{}
		v.ValidateList(context.Background(), validator.ListRequest{Path: path.Root("claim_validation_rule"), ConfigValue: validRules}, result)
		ruleDiags.Append(result.Diagnostics...)
	}
	if ruleDiags.HasError() {
		t.Errorf("rule with colon in required_value must be accepted: %v", ruleDiags)
	}

	scopes := response.Schema.Attributes["console_extra_scopes"].(schema.SetAttribute)
	invalidScopes := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile email")})
	var scopeDiags diag.Diagnostics
	for _, v := range scopes.Validators {
		result := &validator.SetResponse{}
		v.ValidateSet(context.Background(), validator.SetRequest{Path: path.Root("console_extra_scopes"), ConfigValue: invalidScopes}, result)
		scopeDiags.Append(result.Diagnostics...)
	}
	if !scopeDiags.HasError() {
		t.Error("scope with whitespace must be rejected by the schema")
	}
}

func TestExternalAuthProviderCollectionNullElements(t *testing.T) {
	t.Parallel()
	r := New()
	schemaResponse := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResponse)
	for _, tc := range []struct {
		name    string
		unknown bool
	}{
		{"null", false},
		{"unknown", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			value := attr.Value(types.StringNull())
			if tc.unknown {
				value = types.StringUnknown()
			}
			for _, name := range []string{"claim_validation_rule"} {
				attribute := schemaResponse.Schema.Attributes[name].(schema.ListAttribute)
				list := types.ListValueMust(types.StringType, []attr.Value{value})
				var diags diag.Diagnostics
				for _, v := range attribute.Validators {
					result := &validator.ListResponse{}
					v.ValidateList(context.Background(), validator.ListRequest{Path: path.Root(name), ConfigValue: list}, result)
					diags.Append(result.Diagnostics...)
				}
				if diags.HasError() == tc.unknown {
					t.Errorf("%s %s element: diagnostics=%v", name, tc.name, diags)
				}
				if !tc.unknown && len(diags.Errors()) != 1 {
					t.Errorf("%s null element should produce one error, got %v", name, diags)
				}
			}
			audiences := schemaResponse.Schema.Attributes["issuer_audiences"].(schema.SetAttribute)
			audienceSet := types.SetValueMust(types.StringType, []attr.Value{value})
			var audienceDiags diag.Diagnostics
			for _, v := range audiences.Validators {
				result := &validator.SetResponse{}
				v.ValidateSet(context.Background(), validator.SetRequest{
					Path: path.Root("issuer_audiences"), ConfigValue: audienceSet,
				}, result)
				audienceDiags.Append(result.Diagnostics...)
			}
			if audienceDiags.HasError() == tc.unknown || (!tc.unknown && len(audienceDiags.Errors()) != 1) {
				t.Errorf("issuer_audiences %s element: diagnostics=%v", tc.name, audienceDiags)
			}
			attribute := schemaResponse.Schema.Attributes["console_extra_scopes"].(schema.SetAttribute)
			set := types.SetValueMust(types.StringType, []attr.Value{value})
			var diags diag.Diagnostics
			for _, v := range attribute.Validators {
				result := &validator.SetResponse{}
				v.ValidateSet(context.Background(), validator.SetRequest{Path: path.Root("console_extra_scopes"), ConfigValue: set}, result)
				diags.Append(result.Diagnostics...)
			}
			if diags.HasError() == tc.unknown {
				t.Errorf("console_extra_scopes %s element: diagnostics=%v", tc.name, diags)
			}
			if !tc.unknown && len(diags.Errors()) != 1 {
				t.Errorf("console_extra_scopes null element should produce one error, got %v", diags)
			}
		})
	}
}

func TestExternalAuthProviderStateRoundTripAndValidateConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := New()
	schemaResponse := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResponse)

	model := ExternalAuthProviderState{
		Cluster:                   types.StringValue("cluster-id"),
		Name:                      types.StringValue("provider-name"),
		IssuerURL:                 types.StringValue("https://issuer.example.com"),
		IssuerAudiences:           types.SetValueMust(types.StringType, []attr.Value{types.StringValue("console")}),
		IssuerCA:                  types.StringNull(),
		ClaimMappingGroupsClaim:   types.StringValue("groups"),
		ClaimMappingUsernameClaim: types.StringValue("email"),
		ClaimValidationRule:       types.ListNull(types.StringType),
		ConsoleClientID:           types.StringValue("console"),
		ConsoleClientSecret:       types.StringValue(""),
		ExtraScopes:               types.SetNull(types.StringType),
		ID:                        types.StringValue("provider-name"),
	}
	state := tfsdk.State{Schema: schemaResponse.Schema}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("set state: %v", diags)
	}
	var roundTrip ExternalAuthProviderState
	if diags := state.Get(ctx, &roundTrip); diags.HasError() {
		t.Fatalf("get state: %v", diags)
	}
	if !reflect.DeepEqual(model, roundTrip) {
		t.Fatalf("state roundtrip mismatch: got %#v", roundTrip)
	}
	validate := &resource.ValidateConfigResponse{}
	r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: state.Raw},
	}, validate)
	if validate.Diagnostics.HasError() {
		t.Fatalf("valid config rejected: %v", validate.Diagnostics)
	}
}

func TestExternalAuthProviderCrossFieldValidation(t *testing.T) {
	t.Parallel()
	base := func() ExternalAuthProviderState {
		return ExternalAuthProviderState{
			ConsoleClientID:     types.StringNull(),
			ConsoleClientSecret: types.StringNull(),
			ExtraScopes:         types.SetNull(types.StringType),
			IssuerAudiences:     types.SetValueMust(types.StringType, []attr.Value{types.StringValue("console")}),
		}
	}
	for _, tc := range []struct {
		name   string
		modify func(*ExternalAuthProviderState)
		want   string
	}{
		{"no console client", func(_ *ExternalAuthProviderState) {}, ""},
		{"explicitly remove client", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("")
			s.ConsoleClientSecret = types.StringValue("")
		}, ""},
		{"empty secret is validated by the API", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
			s.ConsoleClientSecret = types.StringValue("")
		}, ""},
		{"matching client audience", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
		}, ""},
		{"client ID without secret", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
		}, "console_client_secret"},
		{"missing client audience", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("other")
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
		}, "issuer_audiences"},
		{"secret without client", func(s *ExternalAuthProviderState) {
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
		}, "console_client_id"},
		{"scopes without client", func(s *ExternalAuthProviderState) {
			s.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
		}, "console_client_id"},
		{"scopes with client ID but no secret", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
			s.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
		}, "console_client_secret"},
		{"empty scopes without client", func(s *ExternalAuthProviderState) {
			s.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{})
		}, "console_client_id"},
		{"secret with matching client", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
		}, ""},
		{"scopes with matching client", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
			s.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("profile")})
		}, ""},
		{"scopes cannot accompany client removal", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("")
			s.ConsoleClientSecret = types.StringValue("")
			s.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{})
		}, "console_extra_scopes"},
		{"unknown client", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringUnknown()
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
		}, ""},
		{"unknown audience", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("console")
			s.ConsoleClientSecret = types.StringValue("test-only-placeholder")
			s.IssuerAudiences = types.SetUnknown(types.StringType)
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			config := base()
			tc.modify(&config)
			var diags diag.Diagnostics
			validateCrossFields(context.Background(), config, &diags)
			if tc.want == "" && diags.HasError() {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
			if tc.want != "" && (!diags.HasError() || !strings.Contains(diags[0].Detail(), tc.want)) {
				t.Errorf("expected diagnostic containing %q, got %v", tc.want, diags)
			}
		})
	}
}

func TestExternalAuthProviderConsoleClientPairValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := New()
	schemaResponse := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	secretValidators := schemaResponse.Schema.Attributes["console_client_secret"].(schema.StringAttribute).Validators
	scopeValidators := schemaResponse.Schema.Attributes["console_extra_scopes"].(schema.SetAttribute).Validators
	for _, tc := range []struct {
		name       string
		clientID   types.String
		secret     types.String
		emptyScope bool
		wantErrors int
	}{
		{"secret requires client ID", types.StringNull(), types.StringValue("fixture-placeholder"), false, 1},
		{"unknown client ID defers", types.StringUnknown(), types.StringValue("fixture-placeholder"), false, 0},
		{"known client ID", types.StringValue("console"), types.StringValue("fixture-placeholder"), false, 0},
		{"empty pair removes client", types.StringValue(""), types.StringValue(""), false, 0},
		{"empty secret is delegated to the API", types.StringValue("console"), types.StringValue(""), false, 0},
		{"empty client ID requires empty secret", types.StringValue(""), types.StringNull(), false, 1},
		{"empty pair cannot manage scopes", types.StringValue(""), types.StringValue(""), true, 1},
		{"client ID requires secret", types.StringValue("console"), types.StringNull(), false, 1},
		{"scopes with client ID require secret", types.StringValue("console"), types.StringNull(), true, 1},
		{"null secret", types.StringNull(), types.StringNull(), false, 0},
		{"empty scopes requires client ID", types.StringNull(), types.StringNull(), true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			model := testAuthState(t)
			model.ConsoleClientID = tc.clientID
			model.ConsoleClientSecret = tc.secret
			if tc.emptyScope {
				model.ExtraScopes = types.SetValueMust(types.StringType, []attr.Value{})
			}
			state := tfsdk.State{Schema: schemaResponse.Schema}
			if diags := state.Set(ctx, &model); diags.HasError() {
				t.Fatalf("set test config: %v", diags)
			}
			config := tfsdk.Config{Schema: schemaResponse.Schema, Raw: state.Raw}
			var diags diag.Diagnostics
			for _, v := range secretValidators {
				result := &validator.StringResponse{}
				v.ValidateString(ctx, validator.StringRequest{
					Path:           path.Root("console_client_secret"),
					PathExpression: path.MatchRoot("console_client_secret"),
					Config:         config,
					ConfigValue:    tc.secret,
				}, result)
				diags.Append(result.Diagnostics...)
			}
			if tc.emptyScope {
				for _, v := range scopeValidators {
					result := &validator.SetResponse{}
					v.ValidateSet(ctx, validator.SetRequest{
						Path:           path.Root("console_extra_scopes"),
						PathExpression: path.MatchRoot("console_extra_scopes"),
						Config:         config,
						ConfigValue:    model.ExtraScopes,
					}, result)
					diags.Append(result.Diagnostics...)
				}
			}
			validate := &resource.ValidateConfigResponse{}
			r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: config}, validate)
			diags.Append(validate.Diagnostics...)
			if got := len(diags.Errors()); got != tc.wantErrors {
				t.Errorf("expected %d validation errors, got %d: %v", tc.wantErrors, got, diags)
			}
		})
	}
}

func TestExternalAuthProviderResolvedPlanValidationBeforeOCM(t *testing.T) {
	ctx := context.Background()
	r := &ExternalAuthProviderResource{} // No client: invalid plans must return before any OCM call.
	schemaResponse := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResponse)
	for _, tc := range []struct {
		name   string
		modify func(*ExternalAuthProviderState)
	}{
		{"client ID outside audiences", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringValue("not-in-audiences")
		}},
		{"secret without client ID", func(s *ExternalAuthProviderState) {
			s.ConsoleClientID = types.StringNull()
		}},
		{"client ID without secret", func(s *ExternalAuthProviderState) {
			s.ConsoleClientSecret = types.StringNull()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := testAuthState(t)
			tc.modify(&model)
			state := tfsdk.State{Schema: schemaResponse.Schema}
			if diags := state.Set(ctx, &model); diags.HasError() {
				t.Fatalf("set test state: %v", diags)
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema, Raw: state.Raw}
			createResponse := &resource.CreateResponse{}
			r.Create(ctx, resource.CreateRequest{
				Plan: plan, Config: tfsdk.Config{Raw: plan.Raw, Schema: plan.Schema},
			}, createResponse)
			if !createResponse.Diagnostics.HasError() {
				t.Fatal("Create accepted invalid resolved plan before OCM")
			}
			updateResponse := &resource.UpdateResponse{}
			r.Update(ctx, resource.UpdateRequest{
				Plan: plan, State: state, Config: tfsdk.Config{Raw: plan.Raw, Schema: plan.Schema},
			}, updateResponse)
			if !updateResponse.Diagnostics.HasError() {
				t.Fatal("Update accepted invalid resolved plan before OCM")
			}
		})
	}
}

func TestExternalAuthCheckClusterUsesRawSDK(t *testing.T) {
	cluster, err := cmv1.NewCluster().ID("123").State(cmv1.ClusterStateReady).
		Hypershift(cmv1.NewHypershift().Enabled(true)).
		ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(true)).Build()
	if err != nil {
		t.Fatal(err)
	}
	var clusterJSON strings.Builder
	if err := cmv1.MarshalCluster(cluster, &clusterJSON); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError string
	}{
		{"ready HCP", http.StatusOK, clusterJSON.String(), ""},
		{"API error", http.StatusNotFound, `{"kind":"Error","id":"404","reason":"cluster missing"}`,
			`cannot retrieve cluster "123"`},
		{"empty body", http.StatusOK, "", `OCM returned no cluster details for "123"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			r := &ExternalAuthProviderResource{clustersClient: cmv1.NewClustersClient(
				externalAuthRoundTripFunc(func(request *http.Request) (*http.Response, error) {
					calls++
					if request.Method != http.MethodGet || request.URL.Path != "/api/clusters_mgmt/v1/clusters/123" {
						t.Errorf("unexpected cluster request: %s %s", request.Method, request.URL.Path)
					}
					return &http.Response{
						StatusCode: tc.status,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(tc.body)),
					}, nil
				}), "/api/clusters_mgmt/v1/clusters")}
			err := r.checkCluster(context.Background(), "123")
			if calls != 1 {
				t.Errorf("cluster GET calls = %d, want 1", calls)
			}
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("checkCluster() = %v, want success", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("checkCluster() = %v, want %q", err, tc.wantError)
			}
		})
	}
}

func TestExternalAuthImportID(t *testing.T) {
	for _, tc := range []struct {
		input, cluster, name string
		valid                bool
	}{
		{"123,example", "123", "example", true},
		{"123, example", "123", "example", true},
		{" 123 , example ", "123", "example", true},
		{"123", "", "", false},
		{"123,", "123", "", false},
		{"123,example,extra", "", "", false},
	} {
		cluster, name, valid := parseExternalAuthImportID(tc.input)
		if cluster != tc.cluster || name != tc.name || valid != tc.valid {
			t.Errorf("parseExternalAuthImportID(%q) = %q, %q, %t", tc.input, cluster, name, valid)
		}
	}
}

func TestExternalAuthAPIErrorRedactsOldAndNewSecrets(t *testing.T) {
	message := safeAPIError(errors.New("old-placeholder and new-placeholder were rejected"),
		"old-placeholder", "new-placeholder")
	if strings.Contains(message, "old-placeholder") || strings.Contains(message, "new-placeholder") {
		t.Fatalf("secret escaped redaction: %q", message)
	}
}

type externalAuthRoundTripFunc func(*http.Request) (*http.Response, error)

func (f externalAuthRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func testExternalAuthPatchResponse(t *testing.T, status int, reason string) (*cmv1.ExternalAuthUpdateResponse, error) {
	t.Helper()
	client := cmv1.NewExternalAuthClient(externalAuthRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", request.Method)
		}
		body := ""
		if status >= http.StatusBadRequest {
			body = `{"kind":"Error","reason":"` + reason + `"}`
		}
		return &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	}), "/api/clusters_mgmt/v1/clusters/123/external_auth_config/external_auths/example")
	return client.Update().Body(testRemoteAuth(t)).SendContext(context.Background())
}

func TestExternalAuthPendingDeletionResponseClassification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		reason string
		want   bool
	}{
		{"pending deletion", http.StatusBadRequest,
			"Cluster '123' has an external authentication pending deletion", true},
		{"other bad request", http.StatusBadRequest, "invalid claim mapping", false},
		{"wrong status", http.StatusConflict, "external authentication pending deletion", false},
		{"different case", http.StatusBadRequest, "External authentication pending deletion", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response, err := testExternalAuthPatchResponse(t, tc.status, tc.reason)
			if err == nil {
				t.Fatal("expected SDK error response")
			}
			if got := isPendingDeletionPatchResponse(response); got != tc.want {
				t.Errorf("classification = %t; want %t", got, tc.want)
			}
		})
	}
	if isPendingDeletionPatchResponse(nil) {
		t.Fatal("nil response must not be retried")
	}
}

func TestExternalAuthPendingDeletionPatchRetry(t *testing.T) {
	t.Run("eventual success", func(t *testing.T) {
		calls := 0
		response, err := retryPendingDeletionPatch(context.Background(), func(context.Context) (
			*cmv1.ExternalAuthUpdateResponse, error,
		) {
			calls++
			if calls < 3 {
				return testExternalAuthPatchResponse(t, http.StatusBadRequest,
					"Cluster '123' has an external authentication pending deletion")
			}
			return testExternalAuthPatchResponse(t, http.StatusOK, "")
		}, 0, time.Second, 12)
		if err != nil || response.Status() != http.StatusOK || calls != 3 {
			t.Fatalf("retry result: status=%d, err=%v, calls=%d", response.Status(), err, calls)
		}
	})
	t.Run("bounded persistent error", func(t *testing.T) {
		calls := 0
		_, err := retryPendingDeletionPatch(context.Background(), func(context.Context) (
			*cmv1.ExternalAuthUpdateResponse, error,
		) {
			calls++
			return testExternalAuthPatchResponse(t, http.StatusBadRequest,
				"Cluster '123' has an external authentication pending deletion")
		}, 0, time.Second, pendingDeletionMaxRetries)
		if err == nil || calls != 4 {
			t.Fatalf("persistent error: err=%v, calls=%d; want four attempts", err, calls)
		}
	})
	t.Run("unrelated error", func(t *testing.T) {
		calls := 0
		_, err := retryPendingDeletionPatch(context.Background(), func(context.Context) (
			*cmv1.ExternalAuthUpdateResponse, error,
		) {
			calls++
			return testExternalAuthPatchResponse(t, http.StatusBadRequest, "invalid claim mapping")
		}, 0, time.Second, 12)
		if err == nil || calls != 1 {
			t.Fatalf("unrelated error: err=%v, calls=%d; want one attempt", err, calls)
		}
	})
	t.Run("generic error text", func(t *testing.T) {
		calls := 0
		_, err := retryPendingDeletionPatch(context.Background(), func(context.Context) (
			*cmv1.ExternalAuthUpdateResponse, error,
		) {
			calls++
			return nil, errors.New("external authentication pending deletion")
		}, 0, time.Second, 12)
		if err == nil || calls != 1 {
			t.Fatalf("generic error: err=%v, calls=%d; want one attempt", err, calls)
		}
	})
	t.Run("caller cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		_, err := retryPendingDeletionPatch(ctx, func(context.Context) (
			*cmv1.ExternalAuthUpdateResponse, error,
		) {
			calls++
			response, sendErr := testExternalAuthPatchResponse(t, http.StatusBadRequest,
				"external authentication pending deletion")
			cancel()
			return response, sendErr
		}, time.Hour, time.Minute, 12)
		if !errors.Is(err, context.Canceled) || calls != 1 {
			t.Fatalf("cancellation: err=%v, calls=%d; want one attempt", err, calls)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		calls := 0
		_, err := retryPendingDeletionPatch(context.Background(), func(context.Context) (
			*cmv1.ExternalAuthUpdateResponse, error,
		) {
			calls++
			return testExternalAuthPatchResponse(t, http.StatusBadRequest,
				"external authentication pending deletion")
		}, time.Hour, 10*time.Millisecond, 12)
		if !errors.Is(err, context.DeadlineExceeded) || calls != 1 {
			t.Fatalf("timeout: err=%v, calls=%d; want one attempt", err, calls)
		}
	})
}
