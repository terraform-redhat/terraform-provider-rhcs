package idps

import (
	"fmt"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func GoogleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"client_id": {
			Description: "Client identifier of a registered Google OAuth application.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client_secret": {
			Description: "Client secret issued by Google.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
		"hosted_domain": {
			Description: "Restrict users to a Google Apps domain.",
			Type:        schema.TypeString,
			Optional:    true,
		}}
}

type GoogleIdentityProvider struct {
	// required
	ClientID     string `tfsdk:"client_id"`
	ClientSecret string `tfsdk:"client_secret"`

	// optional
	HostedDomain *string `tfsdk:"hosted_domain"`
}

func ExpandGoogleFromResourceData(resourceData *schema.ResourceData) *GoogleIdentityProvider {
	list, ok := resourceData.GetOk("google")
	if !ok {
		return nil
	}

	return ExpandGoogleFromInterface(list)
}

func ExpandGoogleFromInterface(i interface{}) *GoogleIdentityProvider {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	googleMap := l[0].(map[string]interface{})
	return &GoogleIdentityProvider{
		ClientID:     googleMap["client_id"].(string),
		ClientSecret: googleMap["client_secret"].(string),
		HostedDomain: common2.GetOptionalStringFromMapString(googleMap, "hosted_domain"),
	}
}

func GoogleValidators(i interface{}) error {
	google := ExpandGoogleFromInterface(i)
	if google.HostedDomain != nil && *google.HostedDomain != "" {
		if !common2.IsValidDomain(*google.HostedDomain) {
			return fmt.Errorf(
				fmt.Sprintf("Invalid Google IDP resource configuration. Expected a valid Google hosted_domain. Got %s",
					*google.HostedDomain),
			)
		}
	}
	return nil
}

func CreateGoogleIDPBuilder(mappingMethod string, state *GoogleIdentityProvider) (*cmv1.GoogleIdentityProviderBuilder, error) {
	builder := cmv1.NewGoogleIdentityProvider()
	builder.ClientID(state.ClientID)
	builder.ClientSecret(state.ClientSecret)

	// Mapping method validation. if mappingMethod != lookup, then hosted-domain is mandatory.
	if mappingMethod != string(cmv1.IdentityProviderMappingMethodLookup) {
		if common2.IsStringAttributeEmpty(state.HostedDomain) {
			return nil, fmt.Errorf("Expected a valid hosted_domain since mapping_method is set to %s", mappingMethod)
		}
	}

	if !common2.IsStringAttributeEmpty(state.HostedDomain) {
		builder.HostedDomain(*state.HostedDomain)
	}

	return builder, nil
}

func FlatGoogle(object *cmv1.IdentityProvider) []interface{} {
	gitlabObject, ok := object.GetGoogle()
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	result["client_id"] = gitlabObject.ClientID()
	result["client_secret"] = gitlabObject.ClientSecret()

	if hostedDomain, ok := gitlabObject.GetHostedDomain(); ok {
		result["hosted_domain"] = hostedDomain
	}
	return []interface{}{result}
}
