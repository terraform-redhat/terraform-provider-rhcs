package idps

import (
	"fmt"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func GitlabSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"client_id": {
			Description: "Client identifier of a registered Gitlab OAuth application.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client_secret": {
			Description: "Client secret issued by Gitlab.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
		"url": {
			Description: "URL of the Gitlab instance.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

type GitlabIdentityProvider struct {
	// required
	ClientID     string `tfsdk:"client_id"`
	ClientSecret string `tfsdk:"client_secret"`
	URL          string `tfsdk:"url"`

	// optional
	CA *string `tfsdk:"ca"`
}

func ExpandGitlabFromResourceData(resourceData *schema.ResourceData) *GitlabIdentityProvider {
	list, ok := resourceData.GetOk("gitlab")
	if !ok {
		return nil
	}

	return ExpandGitlabFromInterface(list)
}

func ExpandGitlabFromInterface(i interface{}) *GitlabIdentityProvider {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	gitlabMap := l[0].(map[string]interface{})
	return &GitlabIdentityProvider{
		ClientID:     gitlabMap["client_id"].(string),
		ClientSecret: gitlabMap["client_secret"].(string),
		URL:          gitlabMap["url"].(string),
		CA:           common2.GetOptionalStringFromMapString(gitlabMap, "ca"),
	}
}
func GitlabValidators(i interface{}) error {
	gitlab := ExpandGitlabFromInterface(i)
	if gitlab == nil {
		return nil
	}

	u, err := url.ParseRequestURI(gitlab.URL)
	if err != nil || u.Scheme != "https" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("Invalid Gitlab IDP resource configuration." +
			" Expected a valid GitLab clusterservice URL: to use an https:// scheme, must not have query parameters and not have a fragment.")
	}

	return nil
}

func CreateGitlabIDPBuilder(state *GitlabIdentityProvider) *cmv1.GitlabIdentityProviderBuilder {
	gitlabBuilder := cmv1.NewGitlabIdentityProvider()
	if !common2.IsStringAttributeEmpty(state.CA) {
		gitlabBuilder.CA(*state.CA)
	}
	gitlabBuilder.ClientID(state.ClientID)
	gitlabBuilder.ClientSecret(state.ClientSecret)
	gitlabBuilder.URL(state.URL)
	return gitlabBuilder
}

func FlatGitlab(object *cmv1.IdentityProvider) []interface{} {
	gitlabObject, ok := object.GetGitlab()
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	result["client_id"] = gitlabObject.ClientID()
	result["client_secret"] = gitlabObject.ClientSecret()
	result["url"] = gitlabObject.URL()

	if ca, ok := gitlabObject.GetCA(); ok {
		result["ca"] = ca
	}
	return []interface{}{result}
}
