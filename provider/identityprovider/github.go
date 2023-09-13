package identityprovider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

type GithubIdentityProvider struct {
	CA            types.String `tfsdk:"ca"`
	ClientID      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	Hostname      types.String `tfsdk:"hostname"`
	Organizations types.List   `tfsdk:"organizations"`
	Teams         types.List   `tfsdk:"teams"`
}

var githubSchema = map[string]schema.Attribute{
	"client_id": schema.StringAttribute{
		Description: "Client identifier of a registered Github OAuth application.",
		Required:    true,
	},
	"client_secret": schema.StringAttribute{
		Description: "Client secret issued by Github.",
		Required:    true,
		Sensitive:   true,
	},
	"ca": schema.StringAttribute{
		Description: "Path to PEM-encoded certificate file to use when making requests to the server.",
		Optional:    true,
	},
	"hostname": schema.StringAttribute{
		Description: "Optional domain to use with a hosted instance of GitHub Enterprise.",
		Optional:    true,
	},
	"organizations": schema.ListAttribute{
		Description: "Only users that are members of at least one of the listed organizations will be allowed to log in.",
		ElementType: types.StringType,
		Optional:    true,
	},
	"teams": schema.ListAttribute{
		Description: "Only users that are members of at least one of the listed teams will be allowed to log in. The format is <org>/<team>.",
		ElementType: types.StringType,
		Optional:    true,
	},
}

func githubValidators() []validator.Object {
	errSumm := "Invalid GitHub IDP resource configuration"
	return []validator.Object{
		attrvalidators.NewObjectValidator("GitHub IDP requires either organizations or teams",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, ghState)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				// At only one restriction plan is required
				areTeamsDefined := !ghState.Teams.IsUnknown() && !ghState.Teams.IsNull()
				areOrgsDefined := !ghState.Organizations.IsUnknown() && !ghState.Organizations.IsNull()
				if !areOrgsDefined && !areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires missing attributes 'organizations' OR 'teams'")
				}
				if areOrgsDefined && areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires either 'organizations' or 'teams', not both.")
				}
			}),
		attrvalidators.NewObjectValidator("GitHub IDP teams format validation",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, ghState)
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
			}),
		attrvalidators.NewObjectValidator("GitHub IDP hostname validator",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, ghState)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				// Validate hostname
				if !ghState.Hostname.IsUnknown() && !ghState.Hostname.IsNull() && len(ghState.Hostname.ValueString()) > 0 {
					_, err := url.ParseRequestURI(ghState.Hostname.ValueString())
					if err != nil {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a valid GitHub hostname. Got %v",
								ghState.Hostname.ValueString()),
						)
					}
				}
			}),
	}
}

func CreateGithubIDPBuilder(ctx context.Context, state *GithubIdentityProvider) (*cmv1.GithubIdentityProviderBuilder, error) {
	githubBuilder := cmv1.NewGithubIdentityProvider()
	githubBuilder.ClientID(state.ClientID.ValueString())
	githubBuilder.ClientSecret(state.ClientSecret.ValueString())
	if !state.CA.IsUnknown() && !state.CA.IsNull() {
		githubBuilder.CA(state.CA.ValueString())
	}
	if !state.Hostname.IsUnknown() && !state.Hostname.IsNull() && len(state.Hostname.ValueString()) > 0 {
		githubBuilder.Hostname(state.Hostname.ValueString())
	}
	if !state.Teams.IsUnknown() && !state.Teams.IsNull() {
		teams, err := common.StringListToArray(ctx, state.Teams)
		if err != nil {
			return nil, err
		}
		githubBuilder.Teams(teams...)
	}
	if !state.Organizations.IsUnknown() && !state.Organizations.IsNull() {
		orgs, err := common.StringListToArray(ctx, state.Organizations)
		if err != nil {
			return nil, err
		}
		githubBuilder.Organizations(orgs...)
	}
	return githubBuilder, nil
}
