package idps

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
***REMOVED***

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

func LdapSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"bind_dn": {
			Description: "DN to bind with during the search phase.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"bind_password": {
			Description: "Password to bind with during the search phase.",
			Type:        types.StringType,
			Required:    true,
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
			Attributes:  ldapAttributesSchema(***REMOVED***,
			Required:    true,
***REMOVED***,
	}***REMOVED***
}

func ldapAttributesSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Description: "The list of attributes whose values should be used as the email address.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"id": {
			Description: "The list of attributes whose values should be used as the user ID. (default 'dn'***REMOVED***",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"name": {
			Description: "The list of attributes whose values should be used as the display name. (default 'cn'***REMOVED***",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"preferred_username": {
			Description: "The list of attributes whose values should be used as the preferred username. (default 'uid'***REMOVED***",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
	}***REMOVED***
}

func CreateLdapIDPBuilder(ctx context.Context, state *LDAPIdentityProvider***REMOVED*** (*cmv1.LDAPIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewLDAPIdentityProvider(***REMOVED***
	if !state.BindDN.Null {
		builder.BindDN(state.BindDN.Value***REMOVED***
	}
	if !state.BindPassword.Null {
		builder.BindPassword(state.BindPassword.Value***REMOVED***
	}
	if !state.CA.Null {
		builder.CA(state.CA.Value***REMOVED***
	}
	if !state.Insecure.Null {
		builder.Insecure(state.Insecure.Value***REMOVED***
	}
	if !state.URL.Null {
		builder.URL(state.URL.Value***REMOVED***
	}
	if state.Attributes != nil {
		attributesBuilder := cmv1.NewLDAPAttributes(***REMOVED***
		if !state.Attributes.ID.Unknown && !state.Attributes.ID.Null {
			ids, err := common.StringListToArray(state.Attributes.ID***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			attributesBuilder.ID(ids...***REMOVED***
***REMOVED***
		if !state.Attributes.EMail.Unknown && !state.Attributes.EMail.Null {
			emails, err := common.StringListToArray(state.Attributes.EMail***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			attributesBuilder.Email(emails...***REMOVED***
***REMOVED***
		if !state.Attributes.Name.Unknown && !state.Attributes.Name.Null {
			names, err := common.StringListToArray(state.Attributes.Name***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			attributesBuilder.Name(names...***REMOVED***
***REMOVED***
		if !state.Attributes.PreferredUsername.Unknown && !state.Attributes.PreferredUsername.Null {
			preferredUsernames, err := common.StringListToArray(state.Attributes.PreferredUsername***REMOVED***
			if err != nil {
				return nil, err
	***REMOVED***
			attributesBuilder.PreferredUsername(preferredUsernames...***REMOVED***
***REMOVED***
		builder.Attributes(attributesBuilder***REMOVED***
	}
	return builder, nil
}
