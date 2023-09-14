package identityprovider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const (
	HTPasswdMinPassLength = 14
)

var (
	HTPasswdPassRegexAscii          = regexp.MustCompile(`^[\x20-\x7E]+$`)
	HTPasswdPassRegexHasUpper       = regexp.MustCompile(`[A-Z]`)
	HTPasswdPassRegexHasLower       = regexp.MustCompile(`[a-z]`)
	HTPasswdPassRegexHasNumOrSymbol = regexp.MustCompile(`[^a-zA-Z]`)

	HTPasswdPasswordValidators = []validator.String{
		stringvalidator.LengthAtLeast(HTPasswdMinPassLength),
		stringvalidator.RegexMatches(HTPasswdPassRegexAscii, "password should use ASCII-standard characters only"),
		stringvalidator.RegexMatches(HTPasswdPassRegexHasUpper, "password must contain uppercase characters"),
		stringvalidator.RegexMatches(HTPasswdPassRegexHasLower, "password must contain lowercase characters"),
		stringvalidator.RegexMatches(HTPasswdPassRegexHasNumOrSymbol, "password must contain numbers or symbols"),
	}

	HTPasswdUsernameValidators = []validator.String{
		stringvalidator.RegexMatches(regexp.MustCompile(`^[^/:%]*$`), "username may not contain the characters: '/:%'"),
	}
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
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		Required: true,
	},
}

var htpasswdUserList = map[string]schema.Attribute{
	"username": schema.StringAttribute{
		Description: "User username.",
		Required:    true,
		Validators:  HTPasswdUsernameValidators,
	},
	"password": schema.StringAttribute{
		Description: "User password.",
		Required:    true,
		Sensitive:   true,
		Validators:  HTPasswdPasswordValidators,
	},
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
