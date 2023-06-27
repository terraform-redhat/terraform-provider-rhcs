package idps

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-red-hat-cloud-services/provider/common"
***REMOVED***

var LDAPAttrDefaultID []string = []string{"dn"}
var LDAPAttrDefaultName []string = []string{"cn"}
var LDAPAttrDefaultPrefferedUsername []string = []string{"uid"}

type LDAPIdentityProvider struct {
	BindDN       types.String                    `tfsdk:"bind_dn"`
	BindPassword types.String                    `tfsdk:"bind_password"`
	CA           types.String                    `tfsdk:"ca"`
	Insecure     types.Bool                      `tfsdk:"insecure"`
	URL          types.String                    `tfsdk:"url"`
	Attributes   *LDAPIdentityProviderAttributes `tfsdk:"attributes"`
}

type LDAPIdentityProviderAttributes struct {
	EMail             types.List `tfsdk:"email"`
	ID                types.List `tfsdk:"id"`
	Name              types.List `tfsdk:"name"`
	PreferredUsername types.List `tfsdk:"preferred_username"`
}

func LDAPSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"bind_dn": {
			Description: "DN to bind with during the search phase.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
		"bind_password": {
			Description: "Password to bind with during the search phase.",
			Type:        types.StringType,
			Optional:    true,
			Sensitive:   true,
***REMOVED***,
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
		"insecure": {
			Description: "Do not make TLS connections to the server.",
			Type:        types.BoolType,
			Optional:    true,
			Computed:    true,
***REMOVED***,
		"url": {
			Description: "An RFC 2255 URL which specifies the LDAP search parameters to use.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"attributes": {
			Description: "",
			Attributes:  LDAPAttributesSchema(***REMOVED***,
			Required:    true,
			Validators:  ldapAttrsValidator(***REMOVED***,
***REMOVED***,
	}***REMOVED***
}

func LDAPAttributesSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Description: "The list of attributes whose values should be used as the email address.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"id": {
			Description: "The list of attributes whose values should be used as the user ID. (default ['dn']***REMOVED***",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
			Computed: true,
***REMOVED***,
		"name": {
			Description: "The list of attributes whose values should be used as the display name. (default ['cn']***REMOVED***",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
			Computed: true,
***REMOVED***,
		"preferred_username": {
			Description: "The list of attributes whose values should be used as the preferred username. (default ['uid']***REMOVED***",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
			Computed: true,
***REMOVED***,
	}***REMOVED***
}

func LDAPValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid LDAP IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate bind values",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &LDAPIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				containsBindDN := common.IsStringAttributeEmpty(state.BindDN***REMOVED***
				containsBindPassword := common.IsStringAttributeEmpty(state.BindPassword***REMOVED***
				if containsBindDN != containsBindPassword {
					resp.Diagnostics.AddError(errSumm, "Must provide both `bind_dn` and `bind_password` OR none of them"***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func ldapAttrsValidator(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid LDAP IDP 'attributes' resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate email",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &LDAPIdentityProviderAttributes{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				emails, err := common.StringListToArray(state.EMail***REMOVED***
				if err != nil {
					resp.Diagnostics.AddError(errSumm, "Failed to parse 'email' attribute"***REMOVED***
					return
		***REMOVED***
				for _, email := range emails {
					if !common.IsValidEmail(email***REMOVED*** {
						resp.Diagnostics.AddError(errSumm, fmt.Sprintf("Invalid email '%s'", email***REMOVED******REMOVED***
						return
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func CreateLDAPIDPBuilder(ctx context.Context, state *LDAPIdentityProvider***REMOVED*** (*cmv1.LDAPIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewLDAPIdentityProvider(***REMOVED***
	if !common.IsStringAttributeEmpty(state.BindDN***REMOVED*** {
		builder.BindDN(state.BindDN.Value***REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.BindPassword***REMOVED*** {
		builder.BindPassword(state.BindPassword.Value***REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.CA***REMOVED*** {
		builder.CA(state.CA.Value***REMOVED***
	}
	if !state.Insecure.Null && !state.Insecure.Unknown {
		builder.Insecure(state.Insecure.Value***REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.URL***REMOVED*** {
		builder.URL(state.URL.Value***REMOVED***
	}

	attributesBuilder := cmv1.NewLDAPAttributes(***REMOVED***
	var err error

	var ids []string
	if !state.Attributes.ID.Unknown && !state.Attributes.ID.Null {
		ids, err = common.StringListToArray(state.Attributes.ID***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	} else {
		ids = LDAPAttrDefaultID
		state.Attributes.ID = common.StringArrayToList(ids***REMOVED***
	}
	attributesBuilder.ID(ids...***REMOVED***

	if !state.Attributes.EMail.Unknown && !state.Attributes.EMail.Null {
		emails, err := common.StringListToArray(state.Attributes.EMail***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		attributesBuilder.Email(emails...***REMOVED***
	}

	var names []string
	if !state.Attributes.Name.Unknown && !state.Attributes.Name.Null {
		names, err = common.StringListToArray(state.Attributes.Name***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	} else {
		names = LDAPAttrDefaultName
		state.Attributes.Name = common.StringArrayToList(names***REMOVED***
	}
	attributesBuilder.Name(names...***REMOVED***

	var preferredUsernames []string
	if !state.Attributes.PreferredUsername.Unknown && !state.Attributes.PreferredUsername.Null {
		preferredUsernames, err = common.StringListToArray(state.Attributes.PreferredUsername***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	} else {
		preferredUsernames = LDAPAttrDefaultPrefferedUsername
		state.Attributes.PreferredUsername = common.StringArrayToList(preferredUsernames***REMOVED***
	}
	attributesBuilder.PreferredUsername(preferredUsernames...***REMOVED***

	builder.Attributes(attributesBuilder***REMOVED***
	return builder, nil
}
