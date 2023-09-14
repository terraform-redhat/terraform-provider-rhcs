package identityprovider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

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
			gitlabUrlValidator(),
		},
	},
	"ca": schema.StringAttribute{
		Description: "Optional trusted certificate authority bundle.",
		Optional:    true,
	},
}

func gitlabUrlValidator() validator.String {
	return attrvalidators.NewStringValidator("url validator", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		gitlabUrl := req.ConfigValue
		// Validate hostname
		if !gitlabUrl.IsUnknown() && !gitlabUrl.IsNull() && len(gitlabUrl.ValueString()) > 0 {
			_, err := url.ParseRequestURI(gitlabUrl.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(req.Path, "invalid url",
					fmt.Sprintf("Expected a valid GitLab url. Got %v", gitlabUrl.ValueString()),
				)
			}
		}

	})
}

func CreateGitlabIDPBuilder(ctx context.Context, state *GitlabIdentityProvider) (*cmv1.GitlabIdentityProviderBuilder, error) {
	gitlabBuilder := cmv1.NewGitlabIdentityProvider()
	if !state.CA.IsUnknown() && !state.CA.IsNull() {
		gitlabBuilder.CA(state.CA.ValueString())
	}
	gitlabBuilder.ClientID(state.ClientID.ValueString())
	gitlabBuilder.ClientSecret(state.ClientSecret.ValueString())
	gitlabBuilder.URL(state.URL.ValueString())
	return gitlabBuilder, nil
}
