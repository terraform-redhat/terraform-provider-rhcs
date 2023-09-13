package identityprovider

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
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

var openidSchema = map[string]schema.Attribute{
	"ca": schema.StringAttribute{
		Description: "Optional trusted certificate authority bundle.",
		Optional:    true,
	},
	"claims": schema.SingleNestedAttribute{
		Description: "OpenID Claims config.",
		Attributes:  openidClaimsSchema,
		Required:    true,
	},
	"client_id": schema.StringAttribute{
		Description: "Client ID from the registered application.",
		Required:    true,
	},
	"client_secret": schema.StringAttribute{
		Description: "Client Secret from the registered application.",
		Required:    true,
		Sensitive:   true,
	},
	"extra_scopes": schema.ListAttribute{
		Description: "List of scopes to request, in addition to the 'openid' scope, during the authorization token request.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"extra_authorize_parameters": schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
	},
	"issuer": schema.StringAttribute{
		Description: "The URL that the OpenID Provider asserts as the Issuer Identifier. It must use the https scheme with no URL query parameters or fragment.",
		Required:    true,
	},
}

var openidClaimsSchema = map[string]schema.Attribute{
	"email": schema.ListAttribute{
		Description: "List of claims to use as the email address.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"groups": schema.ListAttribute{
		Description: "List of claims to use as the groups names.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"name": schema.ListAttribute{
		Description: "List of claims to use as the display name.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"preferred_username": schema.ListAttribute{
		Description: "List of claims to use as the preferred username when provisioning a user.",
		ElementType: types.StringType,
		Optional:    true,
	},
}

func CreateOpenIDIDPBuilder(ctx context.Context, state *OpenIDIdentityProvider***REMOVED*** (*cmv1.OpenIDIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewOpenIDIdentityProvider(***REMOVED***
	if !state.CA.IsNull(***REMOVED*** {
		builder.CA(state.CA.ValueString(***REMOVED******REMOVED***
	}
	if state.Claims != nil {
		claimsBuilder := cmv1.NewOpenIDClaims(***REMOVED***

		if !state.Claims.Groups.IsUnknown(***REMOVED*** && !state.Claims.Groups.IsUnknown(***REMOVED*** {
			groups, err := common.StringListToArray(ctx, state.Claims.Groups***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.Groups(groups...***REMOVED***
***REMOVED***
		if !state.Claims.EMail.IsUnknown(***REMOVED*** && !state.Claims.EMail.IsNull(***REMOVED*** {
			emails, err := common.StringListToArray(ctx, state.Claims.EMail***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.Email(emails...***REMOVED***
***REMOVED***
		if !state.Claims.Name.IsUnknown(***REMOVED*** && !state.Claims.Name.IsNull(***REMOVED*** {
			names, err := common.StringListToArray(ctx, state.Claims.Name***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.Name(names...***REMOVED***
***REMOVED***
		if !state.Claims.PreferredUsername.IsUnknown(***REMOVED*** && !state.Claims.PreferredUsername.IsNull(***REMOVED*** {
			usernames, err := common.StringListToArray(ctx, state.Claims.PreferredUsername***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			claimsBuilder.PreferredUsername(usernames...***REMOVED***
***REMOVED***

		builder.Claims(claimsBuilder***REMOVED***
	}
	if !state.ClientID.IsNull(***REMOVED*** {
		builder.ClientID(state.ClientID.ValueString(***REMOVED******REMOVED***
	}
	if !state.ClientSecret.IsNull(***REMOVED*** {
		builder.ClientSecret(state.ClientSecret.ValueString(***REMOVED******REMOVED***
	}
	if state.ExtraAuthorizeParameters != nil {
		builder.ExtraAuthorizeParameters(state.ExtraAuthorizeParameters***REMOVED***
	}
	if !state.ExtraScopes.IsUnknown(***REMOVED*** && !state.ExtraScopes.IsNull(***REMOVED*** {
		extraScopes, err := common.StringListToArray(ctx, state.ExtraScopes***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		builder.ExtraScopes(extraScopes...***REMOVED***
	}
	if !state.Issuer.IsNull(***REMOVED*** {
		builder.Issuer(state.Issuer.ValueString(***REMOVED******REMOVED***
	}
	return builder, nil
}
