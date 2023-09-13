package identityprovider

***REMOVED***
	"context"
***REMOVED***
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
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
		Validators: htpasswdListValidators(***REMOVED***,
		Required:   true,
	},
}

var htpasswdUserList = map[string]schema.Attribute{
	"username": schema.StringAttribute{
		Description: "User username.",
		Required:    true,
		Validators:  htpasswdUsernameValidators(***REMOVED***,
	},
	"password": schema.StringAttribute{
		Description: "User password.",
		Required:    true,
		Sensitive:   true,
		Validators:  htpasswdPasswordValidators(***REMOVED***,
	},
}

func htpasswdListValidators(***REMOVED*** []validator.List {
	return []validator.List{
		attrvalidators.NewListValidator("User list can not be empty",
			func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse***REMOVED*** {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					return
		***REMOVED***
				if len(state.Users***REMOVED*** < 1 {
					resp.Diagnostics.AddAttributeError(req.Path, "user list must contain at least one user object", ""***REMOVED***
					return
		***REMOVED***
	***REMOVED******REMOVED***,
	}
}

func htpasswdUsernameValidators(***REMOVED*** []validator.String {
	return []validator.String{
		attrvalidators.NewStringValidator("Validate username",
			func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
				username := req.ConfigValue
				if err := ValidateHTPasswdUsername(username.ValueString(***REMOVED******REMOVED***; err != nil {
					resp.Diagnostics.AddAttributeError(req.Path, "invalid username", err.Error(***REMOVED******REMOVED***
					return
		***REMOVED***
	***REMOVED******REMOVED***,
	}
}

func htpasswdPasswordValidators(***REMOVED*** []validator.String {
	return []validator.String{
		attrvalidators.NewStringValidator("Validate password",
			func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
				password := req.ConfigValue
				if err := ValidateHTPasswdPassword(password.ValueString(***REMOVED******REMOVED***; err != nil {
					resp.Diagnostics.AddAttributeError(req.Path, "invalid password", err.Error(***REMOVED******REMOVED***
					return
		***REMOVED***
	***REMOVED******REMOVED***,
	}
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

func ValidateHTPasswdUsername(username string***REMOVED*** error {
	if strings.ContainsAny(username, "/:%"***REMOVED*** {
		return fmt.Errorf("invalid username '%s': "+
			"username must not contain /, :, or %%", username***REMOVED***
	}
	return nil
}

func ValidateHTPasswdPassword(password string***REMOVED*** error {
	notAsciiOnly, _ := regexp.MatchString(`[^\x20-\x7E]`, password***REMOVED***
	containsSpace := strings.Contains(password, " "***REMOVED***
	tooShort := len(password***REMOVED*** < 14
	if notAsciiOnly || containsSpace || tooShort {
		return fmt.Errorf(
			"password must be at least 14 characters (ASCII-standard***REMOVED*** without whitespaces"***REMOVED***
	}
	hasUppercase, _ := regexp.MatchString(`[A-Z]`, password***REMOVED***
	hasLowercase, _ := regexp.MatchString(`[a-z]`, password***REMOVED***
	hasNumberOrSymbol, _ := regexp.MatchString(`[^a-zA-Z]`, password***REMOVED***
	if !hasUppercase || !hasLowercase || !hasNumberOrSymbol {
		return fmt.Errorf(
			"password must include uppercase letters, lowercase letters, and numbers " +
				"or symbols (ASCII-standard characters only***REMOVED***"***REMOVED***
	}
	return nil
}
