package identityprovider

***REMOVED***
	"context"
***REMOVED***
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

type GitlabIdentityProvider struct {
	CA           types.String `tfsdk:"ca"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	URL          types.String `tfsdk:"url"`
}

var gitlabSchema = map[string]schema.Attribute{
	"client_id": schema.StringAttribute{
		Description: "Client identifier of a registered Gitlab OAuth application.",
		Required:    true,
	},
	"client_secret": schema.StringAttribute{
		Description: "Client secret issued by Gitlab.",
		Required:    true,
		Sensitive:   true,
	},
	"url": schema.StringAttribute{
		Description: "URL of the Gitlab instance.",
		Required:    true,
		Validators: []validator.String{
			gitlabUrlValidator(***REMOVED***,
***REMOVED***,
	},
	"ca": schema.StringAttribute{
		Description: "Optional trusted certificate authority bundle.",
		Optional:    true,
	},
}

func gitlabUrlValidator(***REMOVED*** validator.String {
	return attrvalidators.NewStringValidator("url validator", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
		gitlabUrl := req.ConfigValue
		// Validate hostname
		if !gitlabUrl.IsUnknown(***REMOVED*** && !gitlabUrl.IsNull(***REMOVED*** && len(gitlabUrl.ValueString(***REMOVED******REMOVED*** > 0 {
			_, err := url.ParseRequestURI(gitlabUrl.ValueString(***REMOVED******REMOVED***
			if err != nil {
				resp.Diagnostics.AddAttributeError(req.Path, "invalid url",
					fmt.Sprintf("Expected a valid GitLab url. Got %v", gitlabUrl.ValueString(***REMOVED******REMOVED***,
				***REMOVED***
	***REMOVED***
***REMOVED***

	}***REMOVED***
}

func CreateGitlabIDPBuilder(ctx context.Context, state *GitlabIdentityProvider***REMOVED*** (*cmv1.GitlabIdentityProviderBuilder, error***REMOVED*** {
	gitlabBuilder := cmv1.NewGitlabIdentityProvider(***REMOVED***
	if !state.CA.IsUnknown(***REMOVED*** && !state.CA.IsNull(***REMOVED*** {
		gitlabBuilder.CA(state.CA.ValueString(***REMOVED******REMOVED***
	}
	gitlabBuilder.ClientID(state.ClientID.ValueString(***REMOVED******REMOVED***
	gitlabBuilder.ClientSecret(state.ClientSecret.ValueString(***REMOVED******REMOVED***
	gitlabBuilder.URL(state.URL.ValueString(***REMOVED******REMOVED***
	return gitlabBuilder, nil
}
