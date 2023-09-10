package idps

***REMOVED***
	"context"
***REMOVED***
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type GithubIdentityProvider struct {
	CA            types.String `tfsdk:"ca"`
	ClientID      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	Hostname      types.String `tfsdk:"hostname"`
	Organizations types.List   `tfsdk:"organizations"`
	Teams         types.List   `tfsdk:"teams"`
}

func GithubSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"client_id": {
			Description: "Client identifier of a registered Github OAuth application.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"client_secret": {
			Description: "Client secret issued by Github.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
		"ca": {
			Description: "Path to PEM-encoded certificate file to use when making requests to the server.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
		"hostname": {
			Description: "Optional domain to use with a hosted instance of GitHub Enterprise.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
		"organizations": {
			Description: "Only users that are members of at least one of the listed organizations will be allowed to log in.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"teams": {
			Description: "Only users that are members of at least one of the listed teams will be allowed to log in. The format is <org>/<team>.",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
	}***REMOVED***
}

func GithubValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid GitHub IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "GitHub IDP requires either organizations or teams",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, ghState***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				// At only one restriction plan is required
				areTeamsDefined := !ghState.Teams.Unknown && !ghState.Teams.Null
				areOrgsDefined := !ghState.Organizations.Unknown && !ghState.Organizations.Null
				if !areOrgsDefined && !areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires missing attributes 'organizations' OR 'teams'"***REMOVED***
		***REMOVED***
				if areOrgsDefined && areTeamsDefined {
					resp.Diagnostics.AddError(errSumm, "GitHub IDP requires either 'organizations' or 'teams', not both."***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
		&common.AttributeValidator{
			Desc: "GitHub IDP teams format validator",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, ghState***REMOVED***
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
	***REMOVED***,
***REMOVED***,
		&common.AttributeValidator{
			Desc: "Github IDP hostname validator",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				ghState := &GithubIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, ghState***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				// Validate hostname
				if !ghState.Hostname.Unknown && !ghState.Hostname.Null && len(ghState.Hostname.Value***REMOVED*** > 0 {
					_, err := url.ParseRequestURI(ghState.Hostname.Value***REMOVED***
					if err != nil {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a valid GitHub hostname. Got %v",
								ghState.Hostname.Value***REMOVED***,
						***REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func CreateGithubIDPBuilder(ctx context.Context, state *GithubIdentityProvider***REMOVED*** (*cmv1.GithubIdentityProviderBuilder, error***REMOVED*** {
	githubBuilder := cmv1.NewGithubIdentityProvider(***REMOVED***
	githubBuilder.ClientID(state.ClientID.Value***REMOVED***
	githubBuilder.ClientSecret(state.ClientSecret.Value***REMOVED***
	if !state.CA.Unknown && !state.CA.Null {
		githubBuilder.CA(state.CA.Value***REMOVED***
	}
	if !state.Hostname.Unknown && !state.Hostname.Null && len(state.Hostname.Value***REMOVED*** > 0 {
		githubBuilder.Hostname(state.Hostname.Value***REMOVED***
	}
	if !state.Teams.Unknown && !state.Teams.Null {
		teams, err := common.StringListToArray(state.Teams***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		githubBuilder.Teams(teams...***REMOVED***
	}
	if !state.Organizations.Unknown && !state.Organizations.Null {
		orgs, err := common.StringListToArray(state.Organizations***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		githubBuilder.Organizations(orgs...***REMOVED***
	}
	return githubBuilder, nil
}
