package idps

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func OpenidSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"claims": {
			Description: "OpenID Claims config.",
			Type:        schema.TypeList,
			MaxItems:    1,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: openidClaimsSchema(),
			},
			Required: true,
		},
		"client_id": {
			Description: "Client ID from the registered application.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client_secret": {
			Description: "Client Secret from the registered application.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
		"extra_scopes": {
			Description: "List of scopes to request, in addition to the 'openid' scope, during the authorization token request.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"extra_authorize_parameters": {
			Type:     schema.TypeMap,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		"issuer": {
			Description: "The URL that the OpenID Provider asserts as the Issuer Identifier. It must use the https scheme with no URL query parameters or fragment.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}
func openidClaimsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"email": {
			Description: "List of claims to use as the email address.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"groups": {
			Description: "List of claims to use as the groups names.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"name": {
			Description: "List of claims to use as the display name.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"preferred_username": {
			Description: "List of claims to use as the preferred username when provisioning a user.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
	}
}

type OpenIDIdentityProvider struct {
	// required
	Claims       OpenIDIdentityProviderClaims `tfsdk:"claims"`
	ClientID     string                       `tfsdk:"client_id"`
	ClientSecret string                       `tfsdk:"client_secret"`
	Issuer       string                       `tfsdk:"issuer"`

	// optional
	CA                       *string           `tfsdk:"ca"`
	ExtraScopes              []string          `tfsdk:"extra_scopes"`
	ExtraAuthorizeParameters map[string]string `tfsdk:"extra_authorize_parameters"`
}

type OpenIDIdentityProviderClaims struct {
	// optional
	EMail             []string `tfsdk:"email"`
	Groups            []string `tfsdk:"groups"`
	Name              []string `tfsdk:"name"`
	PreferredUsername []string `tfsdk:"preferred_username"`
}

func ExpandOpenIDFromResourceData(resourceData *schema.ResourceData) *OpenIDIdentityProvider {
	list, ok := resourceData.GetOk("openid")
	if !ok {
		return nil
	}
	l := list.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	openidMap := l[0].(map[string]interface{})
	return &OpenIDIdentityProvider{
		Claims:       expandOpenIDIdentityProviderClaims(openidMap["claims"].([]interface{})),
		ClientID:     openidMap["client_id"].(string),
		ClientSecret: openidMap["client_secret"].(string),
		Issuer:       openidMap["issuer"].(string),

		CA:                       common2.GetOptionalStringFromMapString(openidMap, "ca"),
		ExtraScopes:              common2.GetOptionalListOfValueStrings(openidMap, "extra_scopes"),
		ExtraAuthorizeParameters: common2.GetOptionalMapString(openidMap, "extra_authorize_parameters"),
	}
}

func expandOpenIDIdentityProviderClaims(l []interface{}) OpenIDIdentityProviderClaims {
	if len(l) == 0 || l[0] == nil {
		return OpenIDIdentityProviderClaims{}
	}
	claimsMap := l[0].(map[string]interface{})
	return OpenIDIdentityProviderClaims{
		EMail:             common2.GetOptionalListOfValueStrings(claimsMap, "email"),
		Groups:            common2.GetOptionalListOfValueStrings(claimsMap, "groups"),
		Name:              common2.GetOptionalListOfValueStrings(claimsMap, "name"),
		PreferredUsername: common2.GetOptionalListOfValueStrings(claimsMap, "preferred_username"),
	}
}

func CreateOpenIDIDPBuilder(state *OpenIDIdentityProvider) *cmv1.OpenIDIdentityProviderBuilder {
	builder := cmv1.NewOpenIDIdentityProvider()
	if !common2.IsStringAttributeEmpty(state.CA) {
		builder.CA(*state.CA)
	}

	claimsBuilder := cmv1.NewOpenIDClaims()

	if !common2.IsListAttributeEmpty(state.Claims.Groups) {
		claimsBuilder.Groups(state.Claims.Groups...)
	}
	if !common2.IsListAttributeEmpty(state.Claims.EMail) {
		claimsBuilder.Email(state.Claims.EMail...)
	}
	if !common2.IsListAttributeEmpty(state.Claims.Name) {
		claimsBuilder.Name(state.Claims.Name...)
	}
	if !common2.IsListAttributeEmpty(state.Claims.PreferredUsername) {
		claimsBuilder.PreferredUsername(state.Claims.PreferredUsername...)
	}

	builder.Claims(claimsBuilder)

	builder.ClientID(state.ClientID)
	builder.ClientSecret(state.ClientSecret)
	builder.Issuer(state.Issuer)

	if state.ExtraAuthorizeParameters != nil {
		builder.ExtraAuthorizeParameters(state.ExtraAuthorizeParameters)
	}
	if !common2.IsListAttributeEmpty(state.ExtraScopes) {
		builder.ExtraScopes(state.ExtraScopes...)
	}
	return builder
}

func FlatOpenID(object *cmv1.IdentityProvider) []interface{} {
	openidObject, ok := object.GetOpenID()
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	result["issuer"] = openidObject.Issuer()
	result["client_id"] = openidObject.ClientID()
	result["client_secret"] = openidObject.ClientSecret()

	if ca, ok := openidObject.GetCA(); ok {
		result["ca"] = ca
	}

	if extraScopes, ok := openidObject.GetExtraScopes(); ok {
		result["extra_scopes"] = extraScopes
	}
	if extraAuth, ok := openidObject.GetExtraAuthorizeParameters(); ok {
		result["extra_authorize_parameters"] = extraAuth
	}

	// attributes:
	claims := openidObject.Claims()
	if claims == nil {
		result["claims"] = []interface{}{}
	} else {
		claimsMap := make(map[string]interface{})
		claimsMap["email"] = claims.Email()
		claimsMap["groups"] = claims.Groups()
		claimsMap["name"] = claims.Name()
		claimsMap["preferred_username"] = claims.PreferredUsername()

		result["attributes"] = []interface{}{claimsMap}
	}
	return []interface{}{result}
}
