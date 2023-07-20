package idps

***REMOVED***
	"context"
***REMOVED***
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
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

func HTPasswdValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid HTPasswd IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate username",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				if !common.IsStringAttributeEmpty(state.Username***REMOVED*** {
					if err := ValidateHTPasswdUsername(state.Username.Value***REMOVED***; err != nil {
						resp.Diagnostics.AddError(errSumm, err.Error(***REMOVED******REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
		&common.AttributeValidator{
			Desc: "Validate password",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &HTPasswdIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				if !common.IsStringAttributeEmpty(state.Password***REMOVED*** {
					if err := ValidateHTPasswdPassword(state.Password.Value***REMOVED***; err != nil {
						resp.Diagnostics.AddError(errSumm, err.Error(***REMOVED******REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
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
