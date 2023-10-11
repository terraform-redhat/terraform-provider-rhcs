package identityprovider

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
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
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("bind_password"***REMOVED******REMOVED***,
***REMOVED***,
	},
	"bind_password": schema.StringAttribute{
		Description: "Password to bind with during the search phase.",
		Optional:    true,
		Sensitive:   true,
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("bind_dn"***REMOVED******REMOVED***,
***REMOVED***,
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
	},
}

var ldapAttrSchema = map[string]schema.Attribute{
	"email": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the email address.",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
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
