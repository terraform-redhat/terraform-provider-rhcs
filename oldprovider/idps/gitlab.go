package idps

***REMOVED***
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type GitlabIdentityProvider struct {
	CA           types.String `tfsdk:"ca"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	URL          types.String `tfsdk:"url"`
}

func GitlabSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"client_id": {
			Description: "Client identifier of a registered Gitlab OAuth application.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"client_secret": {
			Description: "Client secret issued by Gitlab.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
		"url": {
			Description: "URL of the Gitlab instance.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
	}***REMOVED***
}

func GitlabValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid Gitlab IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate gitlab 'url'",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &GitlabIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***
				u, err := url.ParseRequestURI(state.URL.Value***REMOVED***
				if err != nil || u.Scheme != "https" || u.RawQuery != "" || u.Fragment != "" {
					resp.Diagnostics.AddError(errSumm,
						"Expected a valid GitLab provider URL: to use an https:// scheme, must not have query parameters and not have a fragment."***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func CreateGitlabIDPBuilder(ctx context.Context, state *GitlabIdentityProvider***REMOVED*** (*cmv1.GitlabIdentityProviderBuilder, error***REMOVED*** {
	gitlabBuilder := cmv1.NewGitlabIdentityProvider(***REMOVED***
	if !state.CA.Unknown && !state.CA.Null {
		gitlabBuilder.CA(state.CA.Value***REMOVED***
	}
	gitlabBuilder.ClientID(state.ClientID.Value***REMOVED***
	gitlabBuilder.ClientSecret(state.ClientSecret.Value***REMOVED***
	gitlabBuilder.URL(state.URL.Value***REMOVED***
	return gitlabBuilder, nil
}
