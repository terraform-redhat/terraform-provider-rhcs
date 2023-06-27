package idps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

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

func LDAPSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"bind_dn": {
			Description: "DN to bind with during the search phase.",
			Type:        types.StringType,
			Optional:    true,
		},
		"bind_password": {
			Description: "Password to bind with during the search phase.",
			Type:        types.StringType,
			Optional:    true,
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
			Attributes:  LDAPAttributesSchema(),
			Required:    true,
			Validators:  ldapAttrsValidator(),
		},
	})
}

func LDAPAttributesSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Description: "The list of attributes whose values should be used as the email address.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"id": {
			Description: "The list of attributes whose values should be used as the user ID. (default ['dn'])",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			Computed: true,
		},
		"name": {
			Description: "The list of attributes whose values should be used as the display name. (default ['cn'])",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			Computed: true,
		},
		"preferred_username": {
			Description: "The list of attributes whose values should be used as the preferred username. (default ['uid'])",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			Computed: true,
		},
	})
}

func LDAPValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid LDAP IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate bind values",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &LDAPIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				containsBindDN := common.IsStringAttributeEmpty(state.BindDN)
				containsBindPassword := common.IsStringAttributeEmpty(state.BindPassword)
				if containsBindDN != containsBindPassword {
					resp.Diagnostics.AddError(errSumm, "Must provide both `bind_dn` and `bind_password` OR none of them")
				}
			},
		},
	}
}

func ldapAttrsValidator() []tfsdk.AttributeValidator {
	errSumm := "Invalid LDAP IDP 'attributes' resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate email",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &LDAPIdentityProviderAttributes{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				emails, err := common.StringListToArray(state.EMail)
				if err != nil {
					resp.Diagnostics.AddError(errSumm, "Failed to parse 'email' attribute")
					return
				}
				for _, email := range emails {
					if !common.IsValidEmail(email) {
						resp.Diagnostics.AddError(errSumm, fmt.Sprintf("Invalid email '%s'", email))
						return
					}
				}
			},
		},
	}
}

func CreateLDAPIDPBuilder(ctx context.Context, state *LDAPIdentityProvider) (*cmv1.LDAPIdentityProviderBuilder, error) {
	builder := cmv1.NewLDAPIdentityProvider()
	if !common.IsStringAttributeEmpty(state.BindDN) {
		builder.BindDN(state.BindDN.Value)
	}
	if !common.IsStringAttributeEmpty(state.BindPassword) {
		builder.BindPassword(state.BindPassword.Value)
	}
	if !common.IsStringAttributeEmpty(state.CA) {
		builder.CA(state.CA.Value)
	}
	if !state.Insecure.Null && !state.Insecure.Unknown {
		builder.Insecure(state.Insecure.Value)
	}
	if !common.IsStringAttributeEmpty(state.URL) {
		builder.URL(state.URL.Value)
	}

	attributesBuilder := cmv1.NewLDAPAttributes()
	var err error

	var ids []string
	if !state.Attributes.ID.Unknown && !state.Attributes.ID.Null {
		ids, err = common.StringListToArray(state.Attributes.ID)
		if err != nil {
			return nil, err
		}
	} else {
		ids = LDAPAttrDefaultID
		state.Attributes.ID = common.StringArrayToList(ids)
	}
	attributesBuilder.ID(ids...)

	if !state.Attributes.EMail.Unknown && !state.Attributes.EMail.Null {
		emails, err := common.StringListToArray(state.Attributes.EMail)
		if err != nil {
			return nil, err
		}
		attributesBuilder.Email(emails...)
	}

	var names []string
	if !state.Attributes.Name.Unknown && !state.Attributes.Name.Null {
		names, err = common.StringListToArray(state.Attributes.Name)
		if err != nil {
			return nil, err
		}
	} else {
		names = LDAPAttrDefaultName
		state.Attributes.Name = common.StringArrayToList(names)
	}
	attributesBuilder.Name(names...)

	var preferredUsernames []string
	if !state.Attributes.PreferredUsername.Unknown && !state.Attributes.PreferredUsername.Null {
		preferredUsernames, err = common.StringListToArray(state.Attributes.PreferredUsername)
		if err != nil {
			return nil, err
		}
	} else {
		preferredUsernames = LDAPAttrDefaultPrefferedUsername
		state.Attributes.PreferredUsername = common.StringArrayToList(preferredUsernames)
	}
	attributesBuilder.PreferredUsername(preferredUsernames...)

	builder.Attributes(attributesBuilder)
	return builder, nil
}
