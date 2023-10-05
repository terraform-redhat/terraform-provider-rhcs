package identityprovider

***REMOVED***
	"context"
***REMOVED***
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
		Validators: []validator.String{
			githubHostnameValidator(***REMOVED***,
***REMOVED***,
	},
	"organizations": schema.ListAttribute{
		Description: "Only users that are members of at least one of the listed organizations will be allowed to log in.",
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.ConflictsWith(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("teams"***REMOVED******REMOVED***,
			listvalidator.ExactlyOneOf(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("teams"***REMOVED***, path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("organizations"***REMOVED******REMOVED***,
***REMOVED***,
	},
	"teams": schema.ListAttribute{
		Description: "Only users that are members of at least one of the listed teams will be allowed to log in. The format is `<org>`/`<team>`.",
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.ConflictsWith(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("organizations"***REMOVED******REMOVED***,
			listvalidator.ExactlyOneOf(path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("teams"***REMOVED***, path.MatchRelative(***REMOVED***.AtParent(***REMOVED***.AtName("organizations"***REMOVED******REMOVED***,
			listvalidator.ValueStringsAre(
				githubTeamsFormatValidator(***REMOVED***,
			***REMOVED***,
***REMOVED***,
	},
}

func githubTeamsFormatValidator(***REMOVED*** validator.String {
	return attrvalidators.NewStringValidator("validate teams format", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
		team := req.ConfigValue
		parts := strings.Split(team.ValueString(***REMOVED***, "/"***REMOVED***
		if len(parts***REMOVED*** != 2 {
			resp.Diagnostics.AddAttributeError(req.Path, "invalid team format",
				fmt.Sprintf("Expected a GitHub team to follow the form '<org>/<team>', Got %s", team.ValueString(***REMOVED******REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}***REMOVED***
}

func githubHostnameValidator(***REMOVED*** validator.String {
	return attrvalidators.NewStringValidator("hostname validator", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
		hostname := req.ConfigValue
		// Validate hostname
		if !hostname.IsUnknown(***REMOVED*** && !hostname.IsNull(***REMOVED*** && len(hostname.ValueString(***REMOVED******REMOVED*** > 0 {
			_, err := url.ParseRequestURI(hostname.ValueString(***REMOVED******REMOVED***
			if err != nil {
				resp.Diagnostics.AddAttributeError(req.Path, "invalid hostname",
					fmt.Sprintf("Expected a valid GitHub hostname. Got %v", hostname.ValueString(***REMOVED******REMOVED***,
				***REMOVED***
	***REMOVED***
***REMOVED***

	}***REMOVED***
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
