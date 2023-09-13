package identityprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
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
		Validators:  ldapAttrsValidators(),
	},
}

var ldapAttrSchema = map[string]schema.Attribute{
	"email": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the email address.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"id": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the user ID. (default ['dn'])",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
	"name": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the display name. (default ['cn'])",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
	"preferred_username": schema.ListAttribute{
		Description: "The list of attributes whose values should be used as the preferred username. (default ['uid'])",
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
	},
}

func LDAPValidators() []validator.Object {
	errSumm := "Invalid LDAP IDP resource configuration"
	return []validator.Object{
		attrvalidators.NewObjectValidator("Validate bind values",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
				state := &LDAPIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				containsBindDN := common.IsStringAttributeEmpty(state.BindDN)
				containsBindPassword := common.IsStringAttributeEmpty(state.BindPassword)
				if containsBindDN != containsBindPassword {
					resp.Diagnostics.AddError(errSumm, "Must provide both `bind_dn` and `bind_password` OR none of them")
				}
			}),
	}
}

func ldapAttrsValidators() []validator.Object {
	errSumm := "Invalid LDAP IDP 'attributes' resource configuration"
	return []validator.Object{
		attrvalidators.NewObjectValidator("Validate email",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
				state := &LDAPIdentityProviderAttributes{}
				diag := req.Config.GetAttribute(ctx, req.Path, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				emails, err := common.StringListToArray(ctx, state.EMail)
				if err != nil {
					resp.Diagnostics.AddError(errSumm, "Failed to parse 'email' attribute")
					return
				}
				for _, email := range emails {
					if !common.IsValidEmail(email) {
						resp.Diagnostics.AddAttributeError(req.Path, errSumm, fmt.Sprintf("Invalid email '%s'", email))
						return
					}
				}
			}),
	}
}

func CreateLDAPIDPBuilder(ctx context.Context, state *LDAPIdentityProvider) (*cmv1.LDAPIdentityProviderBuilder, error) {
	builder := cmv1.NewLDAPIdentityProvider()
	if !common.IsStringAttributeEmpty(state.BindDN) {
		builder.BindDN(state.BindDN.ValueString())
	}
	if !common.IsStringAttributeEmpty(state.BindPassword) {
		builder.BindPassword(state.BindPassword.ValueString())
	}
	if !common.IsStringAttributeEmpty(state.CA) {
		builder.CA(state.CA.ValueString())
	}
	if !state.Insecure.IsNull() && !state.Insecure.IsUnknown() {
		builder.Insecure(state.Insecure.ValueBool())
	}
	if !common.IsStringAttributeEmpty(state.URL) {
		builder.URL(state.URL.ValueString())
	}

	attributesBuilder := cmv1.NewLDAPAttributes()
	var err error

	var ids []string
	if !state.Attributes.ID.IsUnknown() && !state.Attributes.ID.IsNull() {
		ids, err = common.StringListToArray(ctx, state.Attributes.ID)
		if err != nil {
			return nil, err
		}
	} else {
		ids = LDAPAttrDefaultID
		state.Attributes.ID, err = common.StringArrayToList(ids)
		if err != nil {
			return nil, err
		}
	}
	attributesBuilder.ID(ids...)

	if !state.Attributes.EMail.IsUnknown() && !state.Attributes.EMail.IsNull() {
		emails, err := common.StringListToArray(ctx, state.Attributes.EMail)
		if err != nil {
			return nil, err
		}
		attributesBuilder.Email(emails...)
	}

	var names []string
	if !state.Attributes.Name.IsUnknown() && !state.Attributes.Name.IsNull() {
		names, err = common.StringListToArray(ctx, state.Attributes.Name)
		if err != nil {
			return nil, err
		}
	} else {
		names = LDAPAttrDefaultName
		state.Attributes.Name, err = common.StringArrayToList(names)
		if err != nil {
			return nil, err
		}
	}
	attributesBuilder.Name(names...)

	var preferredUsernames []string
	if !state.Attributes.PreferredUsername.IsUnknown() && !state.Attributes.PreferredUsername.IsNull() {
		preferredUsernames, err = common.StringListToArray(ctx, state.Attributes.PreferredUsername)
		if err != nil {
			return nil, err
		}
	} else {
		preferredUsernames = LDAPAttrDefaultPrefferedUsername
		state.Attributes.PreferredUsername, err = common.StringArrayToList(preferredUsernames)
		if err != nil {
			return nil, err
		}
	}
	attributesBuilder.PreferredUsername(preferredUsernames...)

	builder.Attributes(attributesBuilder)
	return builder, nil
}
