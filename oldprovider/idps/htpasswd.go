package idps

***REMOVED***
	"context"
***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type HTPasswdUser struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type HTPasswdIdentityProvider struct {
	Users []HTPasswdUser `tfsdk:"users"`
}

func HtpasswdSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"users": {
			Description: "A list of htpasswd user credentials",
			Attributes:  HTPasswdUserList(***REMOVED***,
			Required:    true,
***REMOVED***,
	}***REMOVED***
}

func HTPasswdUserList(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
		"username": {
			Description: "User username.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"password": {
			Description: "User password.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
	}, tfsdk.ListNestedAttributesOptions{
		MinItems: 1,
	}***REMOVED***
}
func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider***REMOVED*** *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider(***REMOVED***
	userListBuilder := cmv1.NewHTPasswdUserList(***REMOVED***
	userList := []*cmv1.HTPasswdUserBuilder{}
	for _, user := range state.Users {
		userBuilder := &cmv1.HTPasswdUserBuilder{}
		userBuilder.Username(user.Username.Value***REMOVED***
		userBuilder.Password(user.Password.Value***REMOVED***
		userList = append(userList, userBuilder***REMOVED***
	}
	userListBuilder.Items(userList...***REMOVED***
	builder.Users(userListBuilder***REMOVED***
	return builder
}

func HTPasswdValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid HTPasswd IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate users list length",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				if len(state.Users***REMOVED*** < 1 {
					resp.Diagnostics.AddError(errSumm, "Must provide at least one user."***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
		&common.AttributeValidator{
			Desc: "Validate username/password",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				for i, user := range state.Users {
					if err := ValidateHTPasswdUsername(user.Username.Value***REMOVED***; err != nil {
						errMsg := fmt.Sprintf("Invalid username @ index %d. Error: %s", i, err.Error(***REMOVED******REMOVED***
						resp.Diagnostics.AddError(errSumm, errMsg***REMOVED***
						return
			***REMOVED***
					if err := ValidateHTPasswdPassword(user.Password.Value***REMOVED***; err != nil {
						errMsg := fmt.Sprintf("Invalid password @ index %d. Error: %s", i, err.Error(***REMOVED******REMOVED***
						resp.Diagnostics.AddError(errSumm, errMsg***REMOVED***
						return
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}
