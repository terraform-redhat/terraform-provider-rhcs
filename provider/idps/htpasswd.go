package idps

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type HTPasswdIdentityProvider struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func HtpasswdSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"username": {
			Description: "User name.",
			Type:        types.StringType,
			Required:    true,
		},
		"password": {
			Description: "User password.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
	})
}

func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider) *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider()
	if !state.Username.Null {
		builder.Username(state.Username.Value)
	}
	if !state.Password.Null {
		builder.Password(state.Password.Value)
	}
	return builder
}
