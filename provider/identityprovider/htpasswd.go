package identityprovider

***REMOVED***
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

const (
	HTPasswdMinPassLength = 14
***REMOVED***

var (
	HTPasswdPassRegexAscii          = regexp.MustCompile(`^[\x20-\x7E]+$`***REMOVED***
	HTPasswdPassRegexHasUpper       = regexp.MustCompile(`[A-Z]`***REMOVED***
	HTPasswdPassRegexHasLower       = regexp.MustCompile(`[a-z]`***REMOVED***
	HTPasswdPassRegexHasNumOrSymbol = regexp.MustCompile(`[^a-zA-Z]`***REMOVED***

	HTPasswdPasswordValidators = []validator.String{
		stringvalidator.LengthAtLeast(HTPasswdMinPassLength***REMOVED***,
		stringvalidator.RegexMatches(HTPasswdPassRegexAscii, "password should use ASCII-standard characters only"***REMOVED***,
		stringvalidator.RegexMatches(HTPasswdPassRegexHasUpper, "password must contain uppercase characters"***REMOVED***,
		stringvalidator.RegexMatches(HTPasswdPassRegexHasLower, "password must contain lowercase characters"***REMOVED***,
		stringvalidator.RegexMatches(HTPasswdPassRegexHasNumOrSymbol, "password must contain numbers or symbols"***REMOVED***,
	}

	HTPasswdUsernameValidators = []validator.String{
		stringvalidator.RegexMatches(regexp.MustCompile(`^[^/:%]*$`***REMOVED***, "username may not contain the characters: '/:%'"***REMOVED***,
	}
***REMOVED***

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
***REMOVED***,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1***REMOVED***,
***REMOVED***,
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

func CreateHTPasswdIDPBuilder(ctx context.Context, state *HTPasswdIdentityProvider***REMOVED*** *cmv1.HTPasswdIdentityProviderBuilder {
	builder := cmv1.NewHTPasswdIdentityProvider(***REMOVED***
	userListBuilder := cmv1.NewHTPasswdUserList(***REMOVED***
	userList := []*cmv1.HTPasswdUserBuilder{}
	for _, user := range state.Users {
		userBuilder := &cmv1.HTPasswdUserBuilder{}
		userBuilder.Username(user.Username.ValueString(***REMOVED******REMOVED***
		userBuilder.Password(user.Password.ValueString(***REMOVED******REMOVED***
		userList = append(userList, userBuilder***REMOVED***
	}
	userListBuilder.Items(userList...***REMOVED***
	builder.Users(userListBuilder***REMOVED***
	return builder
}
