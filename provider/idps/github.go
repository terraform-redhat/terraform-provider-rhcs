package idps

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
)

type GithubIdentityProvider struct {
	CA            types.String `tfsdk:"ca"`
	ClientID      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	Hostname      types.String `tfsdk:"hostname"`
	Organizations types.List   `tfsdk:"organizations"`
	Teams         types.List   `tfsdk:"teams"`
}

func GithubSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"client_id": {
			Description: "Client identifier of a registered Github OAuth application.",
			Type:        types.StringType,
			Required:    true,
		},
		"client_secret": {
			Description: "Client secret issued by Github.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
		"ca": {
			Description: "Path to PEM-encoded certificate file to use when making requests to the server.",
			Type:        types.StringType,
			Optional:    true,
		},
		"hostname": {
			Description: "Optional domain to use with a hosted instance of GitHub Enterprise.",
			Type:        types.StringType,
			Optional:    true,
		},
		"organizations": {
			Description: "Only users that are members of at least one of the listed organizations will be allowed to log in.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"teams": {
			Description: "Only users that are members of at least one of the listed teams will be allowed to log in. The format is <org>/<team>.",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
	})
}

func GithubValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid GitHub IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc:   "GitHub IDP requires either organizations or teams",
			MDDesc: "",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, ghState)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				// At only one restriction plan is required
				areTeamsDefined := !ghState.Teams.Unknown && !ghState.Teams.Null
				areOrgsDefined := !ghState.Organizations.Unknown && !ghState.Organizations.Null
				if !areOrgsDefined && !areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires missing attributes 'organizations' OR 'teams'")
				}
				if areOrgsDefined && areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires either 'organizations' or 'teams', not both.")
				}
			},
		},
		&common.AttributeValidator{
			Desc:   "GitHub IDP teams format validator",
			MDDesc: "",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, ghState)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				// Validate teams format
				teams := []string{}
				dig := ghState.Teams.ElementsAs(ctx, &teams, false)
				if dig.HasError() {
					// Nothing to validate
					return
				}
				for i, team := range teams {
					parts := strings.Split(team, "/")
					if len(parts) != 2 {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a GitHub team to follow the form '<org>/<team>', Got %v at index %d",
								team, i),
						)
						return
					}
				}
			},
		},
		&common.AttributeValidator{
			Desc:   "Github IDP hostname validator",
			MDDesc: "",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, ghState)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				// Validate hostname
				if !ghState.Hostname.Unknown && !ghState.Hostname.Null && len(ghState.Hostname.Value) > 0 {
					_, err := url.ParseRequestURI(ghState.Hostname.Value)
					if err != nil {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a valid GitHub hostname. Got %v",
								ghState.Hostname.Value),
						)
					}
				}
			},
		},
	}
}

func CreateGithubIDPBuilder(ctx context.Context, ghState *GithubIdentityProvider) (*cmv1.GithubIdentityProviderBuilder, error) {
	githubBuilder := cmv1.NewGithubIdentityProvider()
	githubBuilder.ClientID(ghState.ClientID.Value)
	githubBuilder.ClientSecret(ghState.ClientSecret.Value)
	if !ghState.CA.Unknown && !ghState.CA.Null {
		githubBuilder.CA(ghState.CA.Value)
	}
	if !ghState.Hostname.Unknown && !ghState.Hostname.Null && len(ghState.Hostname.Value) > 0 {
		githubBuilder.Hostname(ghState.Hostname.Value)
	}
	if !ghState.Teams.Unknown && !ghState.Teams.Null {
		teams, err := common.StringListToArray(ghState.Teams)
		if err != nil {
			return nil, err
		}
		githubBuilder.Teams(teams...)
	}
	if !ghState.Organizations.Unknown && !ghState.Organizations.Null {
		orgs, err := common.StringListToArray(ghState.Organizations)
		if err != nil {
			return nil, err
		}
		githubBuilder.Organizations(orgs...)
	}
	return githubBuilder, nil
}
