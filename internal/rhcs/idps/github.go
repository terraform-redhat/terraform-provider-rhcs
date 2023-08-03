package idps

import (
	"fmt"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func GithubSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"client_id": {
			Description: "Client identifier of a registered Github OAuth application.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client_secret": {
			Description: "Client secret issued by Github.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
		"ca": {
			Description: "Path to PEM-encoded certificate file to use when making requests to the server.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"hostname": {
			Description: "Optional domain to use with a hosted instance of GitHub Enterprise.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"organizations": {
			Description: "Only users that are members of at least one of the listed organizations will be allowed to log in.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
		"teams": {
			Description: "Only users that are members of at least one of the listed teams will be allowed to log in. The format is <org>/<team>.",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
		},
	}
}

type GithubIdentityProvider struct {
	ClientID     string `tfsdk:"client_id"`
	ClientSecret string `tfsdk:"client_secret"`

	// optional
	CA            *string  `tfsdk:"ca"`
	Hostname      *string  `tfsdk:"hostname"`
	Organizations []string `tfsdk:"organizations"`
	Teams         []string `tfsdk:"teams"`
}

func ExpandGithubFromResourceData(resourceData *schema.ResourceData) *GithubIdentityProvider {
	list, ok := resourceData.GetOk("github")
	if !ok {
		return nil
	}

	return ExpandGithubFromInterface(list)
}

func ExpandGithubFromInterface(i interface{}) *GithubIdentityProvider {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	githubMap := l[0].(map[string]interface{})
	return &GithubIdentityProvider{
		ClientID:      githubMap["client_id"].(string),
		ClientSecret:  githubMap["client_secret"].(string),
		CA:            common2.GetOptionalStringFromMapString(githubMap, "ca"),
		Hostname:      common2.GetOptionalStringFromMapString(githubMap, "hostname"),
		Organizations: common2.GetOptionalListOfValueStrings(githubMap, "organizations"),
		Teams:         common2.GetOptionalListOfValueStrings(githubMap, "teams"),
	}
}

func GithubValidators(i interface{}) error {
	errSumm := "Invalid GitHub IDP resource configuration. %s"
	github := ExpandGithubFromInterface(i)
	if github == nil {
		return nil
	}

	// At only one restriction plan is required
	areTeamsDefined := github.Teams != nil && len(github.Teams) > 0
	areOrgsDefined := github.Organizations != nil && len(github.Organizations) > 0
	if !areOrgsDefined && !areTeamsDefined {
		return fmt.Errorf(errSumm, "GitHub IDP requires missing attributes 'organizations' OR 'teams'")
	}
	if areOrgsDefined && areTeamsDefined {
		return fmt.Errorf(errSumm, "GitHub IDP requires either 'organizations' or 'teams', not both.")
	}

	// Validate teams format
	for index, team := range github.Teams {
		parts := strings.Split(team, "/")
		if len(parts) != 2 {
			return fmt.Errorf(errSumm,
				fmt.Sprintf("Expected a GitHub team to follow the form '<org>/<team>', Got %s at index %d",
					team, index),
			)
		}
	}

	// Validate hostname
	if github.Hostname != nil && *github.Hostname != "" {
		_, err := url.ParseRequestURI(*github.Hostname)
		if err != nil {
			return fmt.Errorf(errSumm,
				fmt.Sprintf("Expected a valid GitHub hostname. Got %s",
					*github.Hostname),
			)
		}
	}

	return nil
}

func CreateGithubIDPBuilder(state *GithubIdentityProvider) *cmv1.GithubIdentityProviderBuilder {
	githubBuilder := cmv1.NewGithubIdentityProvider()
	githubBuilder.ClientID(state.ClientID)
	githubBuilder.ClientSecret(state.ClientSecret)
	if !common2.IsStringAttributeEmpty(state.CA) {
		githubBuilder.CA(*state.CA)
	}
	if !common2.IsStringAttributeEmpty(state.Hostname) {
		githubBuilder.Hostname(*state.Hostname)
	}
	if !common2.IsListAttributeEmpty(state.Teams) {
		githubBuilder.Teams(state.Teams...)
	}
	if !common2.IsListAttributeEmpty(state.Organizations) {
		githubBuilder.Organizations(state.Organizations...)
	}
	return githubBuilder
}

func FlatGithub(object *cmv1.IdentityProvider) []interface{} {
	gitlabObject, ok := object.GetGithub()

	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	result["client_id"] = gitlabObject.ClientID()
	result["client_secret"] = gitlabObject.ClientSecret()

	if ca, ok := gitlabObject.GetCA(); ok {
		result["ca"] = ca
	}

	if hostname, ok := gitlabObject.GetHostname(); ok {
		result["hostname"] = hostname
	}

	if organizations, ok := gitlabObject.GetOrganizations(); ok {
		result["organizations"] = organizations
	}

	if teams, ok := gitlabObject.GetTeams(); ok {
		result["teams"] = teams
	}
	return []interface{}{result}
}
