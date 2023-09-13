package identityprovider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

type HTPasswdUser struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type HTPasswdIdentityProvider struct {
	Users []HTPasswdUser `tfsdk:"users"`
}

var htpasswdSchema = map[string]schema.Attribute{
	"users": schema.ListNestedAttribute{
		Description: "A list of htpasswd user credentials",
		NestedObject: schema.NestedAttributeObject{
			Attributes: htpasswdUserList,
		},
		Validators: htpasswdListValidators(),
		Required:   true,
	},
}

var htpasswdUserList = map[string]schema.Attribute{
	"username": schema.StringAttribute{
		Description: "User username.",
		Required:    true,
		Validators:  htpasswdUsernameValidators(),
	},
	"password": schema.StringAttribute{
		Description: "User password.",
		Required:    true,
		Sensitive:   true,
		Validators:  htpasswdPasswordValidators(),
	},
}

func htpasswdListValidators() []validator.List {
	return []validator.List{
		attrvalidators.NewListValidator("User list can not be empty",
			func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, state)
				if diag.HasError() {
					return
				}
				if len(state.Users) < 1 {
					resp.Diagnostics.AddAttributeError(req.Path, "user list must contain at least one user object", "")
					return
				}
			}),
	}
}

func htpasswdUsernameValidators() []validator.String {
	return []validator.String{
		attrvalidators.NewStringValidator("Validate username",
			func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
				username := req.ConfigValue
				if err := ValidateHTPasswdUsername(username.ValueString()); err != nil {
					resp.Diagnostics.AddAttributeError(req.Path, "invalid username", err.Error())
					return
				}
			}),
	}
}

func htpasswdPasswordValidators() []validator.String {
	return []validator.String{
		attrvalidators.NewStringValidator("Validate password",
			func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
				password := req.ConfigValue
				if err := ValidateHTPasswdPassword(password.ValueString()); err != nil {
					resp.Diagnostics.AddAttributeError(req.Path, "invalid password", err.Error())
					return
				}
			}),
	}
}

func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider) *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider()
	userListBuilder := cmv1.NewHTPasswdUserList()
	userList := []*cmv1.HTPasswdUserBuilder{}
	for _, user := range state.Users {
		userBuilder := &cmv1.HTPasswdUserBuilder{}
		userBuilder.Username(user.Username.ValueString())
		userBuilder.Password(user.Password.ValueString())
		userList = append(userList, userBuilder)
	}
	userListBuilder.Items(userList...)
	builder.Users(userListBuilder)
	return builder
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
