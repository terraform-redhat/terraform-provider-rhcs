package identityprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

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

func CreateOpenIDIDPBuilder(ctx context.Context, state *OpenIDIdentityProvider) (*cmv1.OpenIDIdentityProviderBuilder, error) {
	builder := cmv1.NewOpenIDIdentityProvider()
	if !state.CA.IsNull() {
		builder.CA(state.CA.ValueString())
	}
	if state.Claims != nil {
		claimsBuilder := cmv1.NewOpenIDClaims()

		if !state.Claims.Groups.IsUnknown() && !state.Claims.Groups.IsUnknown() {
			groups, err := common.StringListToArray(ctx, state.Claims.Groups)
			if err != nil {
				return nil, err
			}
			claimsBuilder.Groups(groups...)
		}
		if !state.Claims.EMail.IsUnknown() && !state.Claims.EMail.IsNull() {
			emails, err := common.StringListToArray(ctx, state.Claims.EMail)
			if err != nil {
				return nil, err
			}
			claimsBuilder.Email(emails...)
		}
		if !state.Claims.Name.IsUnknown() && !state.Claims.Name.IsNull() {
			names, err := common.StringListToArray(ctx, state.Claims.Name)
			if err != nil {
				return nil, err
			}
			claimsBuilder.Name(names...)
		}
		if !state.Claims.PreferredUsername.IsUnknown() && !state.Claims.PreferredUsername.IsNull() {
			usernames, err := common.StringListToArray(ctx, state.Claims.PreferredUsername)
			if err != nil {
				return nil, err
			}
			claimsBuilder.PreferredUsername(usernames...)
		}

		builder.Claims(claimsBuilder)
	}
	if !state.ClientID.IsNull() {
		builder.ClientID(state.ClientID.ValueString())
	}
	if !state.ClientSecret.IsNull() {
		builder.ClientSecret(state.ClientSecret.ValueString())
	}
	if state.ExtraAuthorizeParameters != nil {
		builder.ExtraAuthorizeParameters(state.ExtraAuthorizeParameters)
	}
	if !state.ExtraScopes.IsUnknown() && !state.ExtraScopes.IsNull() {
		extraScopes, err := common.StringListToArray(ctx, state.ExtraScopes)
		if err != nil {
			return nil, err
		}
		builder.ExtraScopes(extraScopes...)
	}
	if !state.Issuer.IsNull() {
		builder.Issuer(state.Issuer.ValueString())
	}
	return builder, nil
}
