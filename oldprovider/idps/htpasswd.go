package idps

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type HTPasswdUser struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type HTPasswdIdentityProvider struct {
	Users []HTPasswdUser `tfsdk:"users"`
}

func HtpasswdSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"users": {
			Description: "A list of htpasswd user credentials",
			Attributes:  HTPasswdUserList(),
			Required:    true,
		},
	})
}

func HTPasswdUserList() tfsdk.NestedAttributes {
	return tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
		"username": {
			Description: "User username.",
			Type:        types.StringType,
			Required:    true,
		},
		"password": {
			Description: "User password.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
	}, tfsdk.ListNestedAttributesOptions{
		MinItems: 1,
	})
}
func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider) *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider()
	userListBuilder := cmv1.NewHTPasswdUserList()
	userList := []*cmv1.HTPasswdUserBuilder{}
	for _, user := range state.Users {
		userBuilder := &cmv1.HTPasswdUserBuilder{}
		userBuilder.Username(user.Username.Value)
		userBuilder.Password(user.Password.Value)
		userList = append(userList, userBuilder)
	}
	userListBuilder.Items(userList...)
	builder.Users(userListBuilder)
	return builder
}

func HTPasswdValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid HTPasswd IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate users list length",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				if len(state.Users) < 1 {
					resp.Diagnostics.AddError(errSumm, "Must provide at least one user.")
				}
			},
		},
		&common.AttributeValidator{
			Desc: "Validate username/password",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				for i, user := range state.Users {
					if err := ValidateHTPasswdUsername(user.Username.Value); err != nil {
						errMsg := fmt.Sprintf("Invalid username @ index %d. Error: %s", i, err.Error())
						resp.Diagnostics.AddError(errSumm, errMsg)
						return
					}
					if err := ValidateHTPasswdPassword(user.Password.Value); err != nil {
						errMsg := fmt.Sprintf("Invalid password @ index %d. Error: %s", i, err.Error())
						resp.Diagnostics.AddError(errSumm, errMsg)
						return
					}
				}
			},
		},
	}
}
