package idps

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type HTPasswdIdentityProvider struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func HtpasswdSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"username": {
			Description: "User name.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"password": {
			Description: "User password.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
	}***REMOVED***
}

func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider***REMOVED*** *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider(***REMOVED***
	if !state.Username.Null {
		builder.Username(state.Username.Value***REMOVED***
	}
	if !state.Password.Null {
		builder.Password(state.Password.Value***REMOVED***
	}
	return builder
}
