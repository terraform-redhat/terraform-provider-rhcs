package identityprovider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		Validators: []validator.String{
			githubHostnameValidator(),
		},
	},
	"organizations": schema.ListAttribute{
		Description: "Only users that are members of at least one of the listed organizations will be allowed to log in.",
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("teams")),
			listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("teams"), path.MatchRelative().AtParent().AtName("organizations")),
		},
	},
	"teams": schema.ListAttribute{
		Description: "Only users that are members of at least one of the listed teams will be allowed to log in. The format is `<org>`/`<team>`.",
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("organizations")),
			listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("teams"), path.MatchRelative().AtParent().AtName("organizations")),
			listvalidator.ValueStringsAre(
				githubTeamsFormatValidator(),
			),
		},
	},
}

func githubTeamsFormatValidator() validator.String {
	return attrvalidators.NewStringValidator("validate teams format", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		team := req.ConfigValue
		parts := strings.Split(team.ValueString(), "/")
		if len(parts) != 2 {
			resp.Diagnostics.AddAttributeError(req.Path, "invalid team format",
				fmt.Sprintf("Expected a GitHub team to follow the form '<org>/<team>', Got %s", team.ValueString()),
			)
			return
		}
	})
}

func githubHostnameValidator() validator.String {
	return attrvalidators.NewStringValidator("hostname validator", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		hostname := req.ConfigValue
		// Validate hostname
		if !hostname.IsUnknown() && !hostname.IsNull() && len(hostname.ValueString()) > 0 {
			_, err := url.ParseRequestURI(hostname.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(req.Path, "invalid hostname",
					fmt.Sprintf("Expected a valid GitHub hostname. Got %v", hostname.ValueString()),
				)
			}
		}

	})
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
