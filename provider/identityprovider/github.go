package identityprovider

***REMOVED***
	"context"
***REMOVED***
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

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

func githubValidators(***REMOVED*** []validator.Object {
	errSumm := "Invalid GitHub IDP resource configuration"
	return []validator.Object{
		attrvalidators.NewObjectValidator("GitHub IDP requires either organizations or teams",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, ghState***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				// At only one restriction plan is required
				areTeamsDefined := !ghState.Teams.IsUnknown(***REMOVED*** && !ghState.Teams.IsNull(***REMOVED***
				areOrgsDefined := !ghState.Organizations.IsUnknown(***REMOVED*** && !ghState.Organizations.IsNull(***REMOVED***
				if !areOrgsDefined && !areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires missing attributes 'organizations' OR 'teams'"***REMOVED***
		***REMOVED***
				if areOrgsDefined && areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires either 'organizations' or 'teams', not both."***REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***,
		attrvalidators.NewObjectValidator("GitHub IDP teams format validation",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, ghState***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				// Validate teams format
				teams := []string{}
				dig := ghState.Teams.ElementsAs(ctx, &teams, false***REMOVED***
				if dig.HasError(***REMOVED*** {
					// Nothing to validate
					return
		***REMOVED***
				for i, team := range teams {
					parts := strings.Split(team, "/"***REMOVED***
					if len(parts***REMOVED*** != 2 {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a GitHub team to follow the form '<org>/<team>', Got %v at index %d",
								team, i***REMOVED***,
						***REMOVED***
						return
			***REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***,
		attrvalidators.NewObjectValidator("GitHub IDP hostname validator",
			func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.Path, ghState***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				// Validate hostname
				if !ghState.Hostname.IsUnknown(***REMOVED*** && !ghState.Hostname.IsNull(***REMOVED*** && len(ghState.Hostname.ValueString(***REMOVED******REMOVED*** > 0 {
					_, err := url.ParseRequestURI(ghState.Hostname.ValueString(***REMOVED******REMOVED***
					if err != nil {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a valid GitHub hostname. Got %v",
								ghState.Hostname.ValueString(***REMOVED******REMOVED***,
						***REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***,
	}
}

func CreateGithubIDPBuilder(ctx context.Context, state *GithubIdentityProvider***REMOVED*** (*cmv1.GithubIdentityProviderBuilder, error***REMOVED*** {
	githubBuilder := cmv1.NewGithubIdentityProvider(***REMOVED***
	githubBuilder.ClientID(state.ClientID.ValueString(***REMOVED******REMOVED***
	githubBuilder.ClientSecret(state.ClientSecret.ValueString(***REMOVED******REMOVED***
	if !state.CA.IsUnknown(***REMOVED*** && !state.CA.IsNull(***REMOVED*** {
		githubBuilder.CA(state.CA.ValueString(***REMOVED******REMOVED***
	}
	if !state.Hostname.IsUnknown(***REMOVED*** && !state.Hostname.IsNull(***REMOVED*** && len(state.Hostname.ValueString(***REMOVED******REMOVED*** > 0 {
		githubBuilder.Hostname(state.Hostname.ValueString(***REMOVED******REMOVED***
	}
	if !state.Teams.IsUnknown(***REMOVED*** && !state.Teams.IsNull(***REMOVED*** {
		teams, err := common.StringListToArray(ctx, state.Teams***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		githubBuilder.Teams(teams...***REMOVED***
	}
	if !state.Organizations.IsUnknown(***REMOVED*** && !state.Organizations.IsNull(***REMOVED*** {
		orgs, err := common.StringListToArray(ctx, state.Organizations***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		githubBuilder.Organizations(orgs...***REMOVED***
	}
	return githubBuilder, nil
}
