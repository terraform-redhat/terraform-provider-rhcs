package idps

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
***REMOVED***

type OpenIDIdentityProvider struct {
	CA                       types.String                  `tfsdk:"ca"`
	Claims                   *OpenIDIdentityProviderClaims `tfsdk:"claims"`
	ClientID                 types.String                  `tfsdk:"client_id"`
	ClientSecret             types.String                  `tfsdk:"client_secret"`
	ExtraScopes              types.List                    `tfsdk:"extra_scopes"`
	ExtraAuthorizeParameters map[string]string             `tfsdk:"extra_authorize_parameters"`
	Issuer                   types.String                  `tfsdk:"issuer"`
}

type OpenIDIdentityProviderClaims struct {
	EMail             types.List `tfsdk:"email"`
	Groups            types.List `tfsdk:"groups"`
	Name              types.List `tfsdk:"name"`
	PreferredUsername types.List `tfsdk:"preferred_username"`
}

func OpenidSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
		"claims": {
			Description: "OpenID Claims config.",
			Attributes:  openidClaimsSchema(***REMOVED***,
			Required:    true,
***REMOVED***,
		"client_id": {
			Description: "Client ID from the registered application.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"client_secret": {
			Description: "Client Secret from the registered application.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
		"extra_scopes": {
			Description: "List of scopes to request, in addition to the 'openid' scope, during the authorization token request.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"extra_authorize_parameters": {
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"issuer": {
			Description: "The URL that the OpenID Provider asserts as the Issuer Identifier. It must use the https scheme with no URL query parameters or fragment.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
	}***REMOVED***
}

func openidClaimsSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Description: "List of claims to use as the email address.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"groups": {
			Description: "List of claims to use as the groups names.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"name": {
			Description: "List of claims to use as the display name.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"preferred_username": {
			Description: "List of claims to use as the preferred username when provisioning a user.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
	}***REMOVED***
}

func CreateOpenIDIDPBuilder(ctx context.Context, state *OpenIDIdentityProvider***REMOVED*** (*cmv1.OpenIDIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewOpenIDIdentityProvider(***REMOVED***
	if !state.CA.Null {
		builder.CA(state.CA.Value***REMOVED***
	}
	if state.Claims != nil {
		claimsBuilder := cmv1.NewOpenIDClaims(***REMOVED***

		if !state.Claims.Groups.Unknown && !state.Claims.Groups.Null {
			groups, err := common.StringListToArray(state.Claims.Groups***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.Groups(groups...***REMOVED***
***REMOVED***
		if !state.Claims.EMail.Unknown && !state.Claims.EMail.Null {
			emails, err := common.StringListToArray(state.Claims.EMail***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.Email(emails...***REMOVED***
***REMOVED***
		if !state.Claims.Name.Unknown && !state.Claims.Name.Null {
			names, err := common.StringListToArray(state.Claims.Name***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.Name(names...***REMOVED***
***REMOVED***
		if !state.Claims.PreferredUsername.Unknown && !state.Claims.PreferredUsername.Null {
			usernames, err := common.StringListToArray(state.Claims.PreferredUsername***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.PreferredUsername(usernames...***REMOVED***
***REMOVED***

		builder.Claims(claimsBuilder***REMOVED***
	}
	if !state.ClientID.Null {
		builder.ClientID(state.ClientID.Value***REMOVED***
	}
	if !state.ClientSecret.Null {
		builder.ClientSecret(state.ClientSecret.Value***REMOVED***
	}
	if state.ExtraAuthorizeParameters != nil {
		builder.ExtraAuthorizeParameters(state.ExtraAuthorizeParameters***REMOVED***
	}
	if !state.ExtraScopes.Unknown && !state.ExtraScopes.Null {
		extraScopes, err := common.StringListToArray(state.ExtraScopes***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		builder.ExtraScopes(extraScopes...***REMOVED***
	}
	if !state.Issuer.Null {
		builder.Issuer(state.Issuer.Value***REMOVED***
	}
	return builder, nil
}
