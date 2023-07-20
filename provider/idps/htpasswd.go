package idps

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
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

func HTPasswdValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid HTPasswd IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate username",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				if !common.IsStringAttributeEmpty(state.Username) {
					if err := ValidateHTPasswdUsername(state.Username.Value); err != nil {
						resp.Diagnostics.AddError(errSumm, err.Error())
					}
				}
			},
		},
		&common.AttributeValidator{
			Desc: "Validate password",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				if !common.IsStringAttributeEmpty(state.Password) {
					if err := ValidateHTPasswdPassword(state.Password.Value); err != nil {
						resp.Diagnostics.AddError(errSumm, err.Error())
					}
				}
			},
		},
	}
}

func ValidateHTPasswdUsername(username string) error {
	if strings.ContainsAny(username, "/:%") {
		return fmt.Errorf("invalid username '%s': "+
			"username must not contain /, :, or %%", username)
	}
	return nil
}

func ValidateHTPasswdPassword(password string) error {
	notAsciiOnly, _ := regexp.MatchString(`[^\x20-\x7E]`, password)
	containsSpace := strings.Contains(password, " ")
	tooShort := len(password) < 14
	if notAsciiOnly || containsSpace || tooShort {
		return fmt.Errorf(
			"password must be at least 14 characters (ASCII-standard) without whitespaces")
	}
	hasUppercase, _ := regexp.MatchString(`[A-Z]`, password)
	hasLowercase, _ := regexp.MatchString(`[a-z]`, password)
	hasNumberOrSymbol, _ := regexp.MatchString(`[^a-zA-Z]`, password)
	if !hasUppercase || !hasLowercase || !hasNumberOrSymbol {
		return fmt.Errorf(
			"password must include uppercase letters, lowercase letters, and numbers " +
				"or symbols (ASCII-standard characters only)")
	}
	return nil
}
