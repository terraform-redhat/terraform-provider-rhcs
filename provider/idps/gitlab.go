package idps

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type GitlabIdentityProvider struct {
	CA           types.String `tfsdk:"ca"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	URL          types.String `tfsdk:"url"`
}

func GitlabSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"client_id": {
			Description: "Client identifier of a registered Gitlab OAuth application.",
			Type:        types.StringType,
			Required:    true,
		},
		"client_secret": {
			Description: "Client secret issued by Gitlab.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
		"url": {
			Description: "URL of the Gitlab instance.",
			Type:        types.StringType,
			Required:    true,
		},
		"ca": {
			Description: "Optional trusted certificate authority bundle.",
			Type:        types.StringType,
			Optional:    true,
		},
	})
}

func GitlabValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid Gitlab IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate gitlab 'url'",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &GitlabIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}
				u, err := url.ParseRequestURI(state.URL.Value)
				if err != nil || u.Scheme != "https" || u.RawQuery != "" || u.Fragment != "" {
					resp.Diagnostics.AddError(errSumm,
						"Expected a valid GitLab provider URL: to use an https:// scheme, must not have query parameters and not have a fragment.")
				}
			},
		},
	}
}

func CreateGitlabIDPBuilder(ctx context.Context, state *GitlabIdentityProvider) (*cmv1.GitlabIdentityProviderBuilder, error) {
	gitlabBuilder := cmv1.NewGitlabIdentityProvider()
	if !state.CA.Unknown && !state.CA.Null {
		gitlabBuilder.CA(state.CA.Value)
	}
	gitlabBuilder.ClientID(state.ClientID.Value)
	gitlabBuilder.ClientSecret(state.ClientSecret.Value)
	gitlabBuilder.URL(state.URL.Value)
	return gitlabBuilder, nil
}
