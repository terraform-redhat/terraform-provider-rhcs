package idps

import (
	"fmt"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

var LDAPAttrDefaultID []string = []string{"dn"}
var LDAPAttrDefaultName []string = []string{"cn"}
var LDAPAttrDefaultPrefferedUsername []string = []string{"uid"}

func LDAPSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bind_dn": {
			Description: "DN to bind with during the search phase.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"bind_password": {
			Description: "Password to bind with during the search phase.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"insecure": {
			Description: "Do not make TLS connections to the server.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"url": {
			Description: "An RFC 2255 URL which specifies the LDAP search parameters to use.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"attributes": {
			Description: "LDAP attributes config",
			Type:        schema.TypeList,
			MaxItems:    1,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: LDAPAttributesSchema(),
			},
			Required: true,
		},
	}
}

func LDAPAttributesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"email": {
			Description: "The list of attributes whose values should be used as the email address.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"id": {
			Description: "The list of attributes whose values should be used as the user ID. (default ['dn'])",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Computed:    true,
		},
		"name": {
			Description: "The list of attributes whose values should be used as the display name. (default ['cn'])",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Computed:    true,
		},
		"preferred_username": {
			Description: "The list of attributes whose values should be used as the preferred username. (default ['uid'])",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Computed:    true,
		},
	}
}

type LDAPIdentityProvider struct {
	// required
	URL        string                         `tfsdk:"url"`
	Attributes LDAPIdentityProviderAttributes `tfsdk:"attributes"`

	// optional
	BindDN       *string `tfsdk:"bind_dn"`
	BindPassword *string `tfsdk:"bind_password"`
	CA           *string `tfsdk:"ca"`
	Insecure     *bool   `tfsdk:"insecure"`
}

type LDAPIdentityProviderAttributes struct {
	EMail             []string `tfsdk:"email"`
	ID                []string `tfsdk:"id"`
	Name              []string `tfsdk:"name"`
	PreferredUsername []string `tfsdk:"preferred_username"`
}

func ExpandLDAPFromResourceData(resourceData *schema.ResourceData) *LDAPIdentityProvider {
	list, ok := resourceData.GetOk("ldap")
	if !ok {
		return nil
	}

	return ExpandLDAPFromInterface(list)
}
func ExpandLDAPFromInterface(i interface{}) *LDAPIdentityProvider {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	ldapMap := l[0].(map[string]interface{})
	return &LDAPIdentityProvider{
		URL:          ldapMap["url"].(string),
		BindDN:       common2.GetOptionalStringFromMapString(ldapMap, "bind_dn"),
		BindPassword: common2.GetOptionalStringFromMapString(ldapMap, "bind_password"),
		CA:           common2.GetOptionalStringFromMapString(ldapMap, "ca"),
		Insecure:     common2.GetOptionalBoolFromMapString(ldapMap, "insecure"),
		Attributes:   expandLDAPIdentityProviderAttributes(ldapMap["attributes"].([]interface{})),
	}
}

func expandLDAPIdentityProviderAttributes(l []interface{}) LDAPIdentityProviderAttributes {
	if len(l) == 0 || l[0] == nil {
		return LDAPIdentityProviderAttributes{}
	}
	attributesMap := l[0].(map[string]interface{})
	return LDAPIdentityProviderAttributes{
		EMail:             common2.GetOptionalListOfValueStrings(attributesMap, "email"),
		ID:                common2.GetOptionalListOfValueStrings(attributesMap, "id"),
		Name:              common2.GetOptionalListOfValueStrings(attributesMap, "name"),
		PreferredUsername: common2.GetOptionalListOfValueStrings(attributesMap, "preferred_username"),
	}
}
func LDAPValidators(i interface{}) error {
	ldap := ExpandLDAPFromInterface(i)

	containsBindDN := !common2.IsStringAttributeEmpty(ldap.BindDN)
	containsBindPassword := !common2.IsStringAttributeEmpty(ldap.BindPassword)
	if containsBindDN != containsBindPassword {
		return fmt.Errorf("Invalid LDAP IDP resource configuration. Must provide both `bind_dn` and `bind_password` OR none of them")
	}

	return nil
}

func ldapAttrsValidator(i interface{}) error {
	ldapAttrs := expandLDAPIdentityProviderAttributes(i.([]interface{}))
	if ldapAttrs.EMail == nil || len(ldapAttrs.EMail) < 1 {
		return nil
	}
	for _, email := range ldapAttrs.EMail {
		if !common2.IsValidEmail(email) {
			return fmt.Errorf(fmt.Sprintf("Invalid LDAP IDP 'attributes' resource configuration. Invalid email '%s'", email))
		}
	}

	return nil
}

func CreateLDAPIDPBuilder(state *LDAPIdentityProvider) *cmv1.LDAPIdentityProviderBuilder {
	builder := cmv1.NewLDAPIdentityProvider()
	if !common2.IsStringAttributeEmpty(state.BindDN) {
		builder.BindDN(*state.BindDN)
	}
	if !common2.IsStringAttributeEmpty(state.BindPassword) {
		builder.BindPassword(*state.BindPassword)
	}
	if !common2.IsStringAttributeEmpty(state.CA) {
		builder.CA(*state.CA)
	}
	if state.Insecure != nil {
		builder.Insecure(*state.Insecure)
	}
	if state.URL != "" {
		builder.URL(state.URL)
	}

	attributesBuilder := cmv1.NewLDAPAttributes()

	var ids []string
	if !common2.IsListAttributeEmpty(state.Attributes.ID) {
		ids = state.Attributes.ID
	} else {
		ids = LDAPAttrDefaultID
		state.Attributes.ID = ids
	}
	attributesBuilder.ID(ids...)

	if !common2.IsListAttributeEmpty(state.Attributes.EMail) {
		attributesBuilder.Email(state.Attributes.EMail...)
	}

	var names []string
	if !common2.IsListAttributeEmpty(state.Attributes.Name) {
		names = state.Attributes.Name
	} else {
		names = LDAPAttrDefaultName
		state.Attributes.Name = names
	}
	attributesBuilder.Name(names...)

	var preferredUsernames []string
	if !common2.IsListAttributeEmpty(state.Attributes.PreferredUsername) {
		preferredUsernames = state.Attributes.PreferredUsername

	} else {
		preferredUsernames = LDAPAttrDefaultPrefferedUsername
		state.Attributes.PreferredUsername = preferredUsernames
	}
	attributesBuilder.PreferredUsername(preferredUsernames...)

	builder.Attributes(attributesBuilder)
	return builder
}

func FlatLDAP(object *cmv1.IdentityProvider) []interface{} {
	ldapObject, ok := object.GetLDAP()
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	result["url"] = ldapObject.URL()

	if bindDN, ok := ldapObject.GetBindDN(); ok {
		result["bind_dn"] = bindDN
	}
	if bindPassword, ok := ldapObject.GetBindPassword(); ok {
		result["bind_password"] = bindPassword
	}
	if ca, ok := ldapObject.GetCA(); ok {
		result["ca"] = ca
	}

	result["insecure"] = ldapObject.Insecure()

	// attributes:
	attributes := ldapObject.Attributes()
	if attributes == nil {
		result["attributes"] = []interface{}{}
	} else {
		attributesMap := make(map[string]interface{})
		attributesMap["email"] = attributes.Email()
		attributesMap["id"] = attributes.ID()
		attributesMap["name"] = attributes.Name()
		attributesMap["preferred_username"] = attributes.PreferredUsername()

		result["attributes"] = []interface{}{attributesMap}
	}
	return []interface{}{result}
}
