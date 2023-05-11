package idps

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
)

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

func LdapSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"bind_dn": {
			Description: "DN to bind with during the search phase.",
			Type:        types.StringType,
			Required:    true,
		},
		"bind_password": {
			Description: "Password to bind with during the search phase.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        types.StringType,
			Optional:    true,
		},
		"insecure": {
			Description: "Do not make TLS connections to the server.",
			Type:        types.BoolType,
			Optional:    true,
			Computed:    true,
		},
		"url": {
			Description: "An RFC 2255 URL which specifies the LDAP search parameters to use.",
			Type:        types.StringType,
			Required:    true,
		},
		"attributes": {
			Description: "",
			Attributes:  ldapAttributesSchema(),
			Required:    true,
		},
	})
}

func ldapAttributesSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Description: "The list of attributes whose values should be used as the email address.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"id": {
			Description: "The list of attributes whose values should be used as the user ID. (default 'dn')",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"name": {
			Description: "The list of attributes whose values should be used as the display name. (default 'cn')",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"preferred_username": {
			Description: "The list of attributes whose values should be used as the preferred username. (default 'uid')",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
	})
}

func CreateLdapIDPBuilder(ctx context.Context, state *LDAPIdentityProvider) (*cmv1.LDAPIdentityProviderBuilder, error) {
	builder := cmv1.NewLDAPIdentityProvider()
	if !state.BindDN.Null {
		builder.BindDN(state.BindDN.Value)
	}
	if !state.BindPassword.Null {
		builder.BindPassword(state.BindPassword.Value)
	}
	if !state.CA.Null {
		builder.CA(state.CA.Value)
	}
	if !state.Insecure.Null {
		builder.Insecure(state.Insecure.Value)
	}
	if !state.URL.Null {
		builder.URL(state.URL.Value)
	}
	if state.Attributes != nil {
		attributesBuilder := cmv1.NewLDAPAttributes()
		if !state.Attributes.ID.Unknown && !state.Attributes.ID.Null {
			ids, err := common.StringListToArray(state.Attributes.ID)
			if err != nil {
				return nil, err
			}
			attributesBuilder.ID(ids...)
		}
		if !state.Attributes.EMail.Unknown && !state.Attributes.EMail.Null {
			emails, err := common.StringListToArray(state.Attributes.EMail)
			if err != nil {
				return nil, err
			}
			attributesBuilder.Email(emails...)
		}
		if !state.Attributes.Name.Unknown && !state.Attributes.Name.Null {
			names, err := common.StringListToArray(state.Attributes.Name)
			if err != nil {
				return nil, err
			}
			attributesBuilder.Name(names...)
		}
		if !state.Attributes.PreferredUsername.Unknown && !state.Attributes.PreferredUsername.Null {
			preferredUsernames, err := common.StringListToArray(state.Attributes.PreferredUsername)
			if err != nil {
				return nil, err
			}
			attributesBuilder.PreferredUsername(preferredUsernames...)
		}
		builder.Attributes(attributesBuilder)
	}
	return builder, nil
}
