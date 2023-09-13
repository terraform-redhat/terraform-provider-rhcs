package identityprovider

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
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

var ldapSchema = map[string]schema.Attribute{
	"bind_dn": schema.StringAttribute{
		Description: "DN to bind with during the search phase.",
		Optional:    true,
	},
	"bind_password": schema.StringAttribute{
		Description: "Password to bind with during the search phase.",
		Optional:    true,
		Sensitive:   true,
	},
	"ca": schema.StringAttribute{
		Description: "Optional trusted certificate authority bundle.",
		Optional:    true,
	},
	"insecure": schema.BoolAttribute{
		Description: "Do not make TLS connections to the server.",
		Optional:    true,
		Computed:    true,
	},
	"url": schema.StringAttribute{
		Description: "An RFC 2255 URL which specifies the LDAP search parameters to use.",
		Required:    true,
	},
	"attributes": schema.SingleNestedAttribute{
		Description: "",
		Attributes:  ldapAttrSchema,
		Required:    true,
		Validators:  ldapAttrsValidators(***REMOVED***,
	},
}

var ldapAttrSchema = map[string]schema.Attribute{
	"email": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the email address.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"id": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the user ID. (default ['dn']***REMOVED***",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
	"name": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the display name. (default ['cn']***REMOVED***",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
	"preferred_username": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the preferred username. (default ['uid']***REMOVED***",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
}

func LDAPValidators(***REMOVED*** []validator.Object {
	errSumm := "Invalid LDAP IDP resource configuration"
	return []validator.Object{
		attrvalidators.NewObjectValidator("Validate bind values",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
				state := &LDAPIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				containsBindDN := common.IsStringAttributeEmpty(state.BindDN***REMOVED***
				containsBindPassword := common.IsStringAttributeEmpty(state.BindPassword***REMOVED***
				if containsBindDN != containsBindPassword {
					resp.Diagnostics.AddError(errSumm, "Must provide both `bind_dn` and `bind_password` OR none of them"***REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***,
	}
}

func ldapAttrsValidators(***REMOVED*** []validator.Object {
	errSumm := "Invalid LDAP IDP 'attributes' resource configuration"
	return []validator.Object{
		attrvalidators.NewObjectValidator("Validate email",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
				state := &LDAPIdentityProviderAttributes{}
				diag := req.Config.GetAttribute(ctx, req.Path, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				emails, err := common.StringListToArray(ctx, state.EMail***REMOVED***
				if err != nil {
					resp.Diagnostics.AddError(errSumm, "Failed to parse 'email' attribute"***REMOVED***
					return
		***REMOVED***
				for _, email := range emails {
					if !common.IsValidEmail(email***REMOVED*** {
						resp.Diagnostics.AddAttributeError(req.Path, errSumm, fmt.Sprintf("Invalid email '%s'", email***REMOVED******REMOVED***
						return
			***REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***,
	}
}

func CreateLDAPIDPBuilder(ctx context.Context, state *LDAPIdentityProvider***REMOVED*** (*cmv1.LDAPIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewLDAPIdentityProvider(***REMOVED***
	if !common.IsStringAttributeEmpty(state.BindDN***REMOVED*** {
		builder.BindDN(state.BindDN.ValueString(***REMOVED******REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.BindPassword***REMOVED*** {
		builder.BindPassword(state.BindPassword.ValueString(***REMOVED******REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.CA***REMOVED*** {
		builder.CA(state.CA.ValueString(***REMOVED******REMOVED***
	}
	if !state.Insecure.IsNull(***REMOVED*** && !state.Insecure.IsUnknown(***REMOVED*** {
		builder.Insecure(state.Insecure.ValueBool(***REMOVED******REMOVED***
	}
	if !common.IsStringAttributeEmpty(state.URL***REMOVED*** {
		builder.URL(state.URL.ValueString(***REMOVED******REMOVED***
	}

	attributesBuilder := cmv1.NewLDAPAttributes(***REMOVED***
	var err error

	var ids []string
	if !state.Attributes.ID.IsUnknown(***REMOVED*** && !state.Attributes.ID.IsNull(***REMOVED*** {
		ids, err = common.StringListToArray(ctx, state.Attributes.ID***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	} else {
		ids = LDAPAttrDefaultID
		state.Attributes.ID, err = common.StringArrayToList(ids***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	}
	attributesBuilder.ID(ids...***REMOVED***

	if !state.Attributes.EMail.IsUnknown(***REMOVED*** && !state.Attributes.EMail.IsNull(***REMOVED*** {
		emails, err := common.StringListToArray(ctx, state.Attributes.EMail***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		attributesBuilder.Email(emails...***REMOVED***
	}

	var names []string
	if !state.Attributes.Name.IsUnknown(***REMOVED*** && !state.Attributes.Name.IsNull(***REMOVED*** {
		names, err = common.StringListToArray(ctx, state.Attributes.Name***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	} else {
		names = LDAPAttrDefaultName
		state.Attributes.Name, err = common.StringArrayToList(names***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	}
	attributesBuilder.Name(names...***REMOVED***

	var preferredUsernames []string
	if !state.Attributes.PreferredUsername.IsUnknown(***REMOVED*** && !state.Attributes.PreferredUsername.IsNull(***REMOVED*** {
		preferredUsernames, err = common.StringListToArray(ctx, state.Attributes.PreferredUsername***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	} else {
		preferredUsernames = LDAPAttrDefaultPrefferedUsername
		state.Attributes.PreferredUsername, err = common.StringArrayToList(preferredUsernames***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
	}
	attributesBuilder.PreferredUsername(preferredUsernames...***REMOVED***

	builder.Attributes(attributesBuilder***REMOVED***
	return builder, nil
}
