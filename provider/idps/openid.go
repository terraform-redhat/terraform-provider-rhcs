package idps

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
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

func OpenidSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        types.StringType,
			Optional:    true,
		},
		"claims": {
			Description: "OpenID Claims config.",
			Attributes:  openidClaimsSchema(),
			Required:    true,
		},
		"client_id": {
			Description: "Client ID from the registered application.",
			Type:        types.StringType,
			Required:    true,
		},
		"client_secret": {
			Description: "Client Secret from the registered application.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
		"extra_scopes": {
			Description: "List of scopes to request, in addition to the 'openid' scope, during the authorization token request.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"extra_authorize_parameters": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"issuer": {
			Description: "The URL that the OpenID Provider asserts as the Issuer Identifier. It must use the https scheme with no URL query parameters or fragment.",
			Type:        types.StringType,
			Required:    true,
		},
	})
}

func openidClaimsSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Description: "List of claims to use as the email address.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"groups": {
			Description: "List of claims to use as the groups names.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"name": {
			Description: "List of claims to use as the display name.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"preferred_username": {
			Description: "List of claims to use as the preferred username when provisioning a user.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
	})
}

func CreateOpenIDIDPBuilder(ctx context.Context, state *OpenIDIdentityProvider) (*cmv1.OpenIDIdentityProviderBuilder, error) {
	builder := cmv1.NewOpenIDIdentityProvider()
	if !state.CA.Null {
		builder.CA(state.CA.Value)
	}
	if state.Claims != nil {
		claimsBuilder := cmv1.NewOpenIDClaims()

		if !state.Claims.Groups.Unknown && !state.Claims.Groups.Null {
			groups, err := common.StringListToArray(state.Claims.Groups)
			if err != nil {
				return nil, err
			}
			claimsBuilder.Groups(groups...)
		}
		if !state.Claims.EMail.Unknown && !state.Claims.EMail.Null {
			emails, err := common.StringListToArray(state.Claims.EMail)
			if err != nil {
				return nil, err
			}
			claimsBuilder.Email(emails...)
		}
		if !state.Claims.Name.Unknown && !state.Claims.Name.Null {
			names, err := common.StringListToArray(state.Claims.Name)
			if err != nil {
				return nil, err
			}
			claimsBuilder.Name(names...)
		}
		if !state.Claims.PreferredUsername.Unknown && !state.Claims.PreferredUsername.Null {
			usernames, err := common.StringListToArray(state.Claims.PreferredUsername)
			if err != nil {
				return nil, err
			}
			claimsBuilder.PreferredUsername(usernames...)
		}

		builder.Claims(claimsBuilder)
	}
	if !state.ClientID.Null {
		builder.ClientID(state.ClientID.Value)
	}
	if !state.ClientSecret.Null {
		builder.ClientSecret(state.ClientSecret.Value)
	}
	if state.ExtraAuthorizeParameters != nil {
		builder.ExtraAuthorizeParameters(state.ExtraAuthorizeParameters)
	}
	if !state.ExtraScopes.Unknown && !state.ExtraScopes.Null {
		extraScopes, err := common.StringListToArray(state.ExtraScopes)
		if err != nil {
			return nil, err
		}
		builder.ExtraScopes(extraScopes...)
	}
	if !state.Issuer.Null {
		builder.Issuer(state.Issuer.Value)
	}
	return builder, nil
}
