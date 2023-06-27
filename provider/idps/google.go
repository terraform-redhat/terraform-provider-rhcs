package idps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-red-hat-cloud-services/provider/common"
)

type GoogleIdentityProvider struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	HostedDomain types.String `tfsdk:"hosted_domain"`
}

func GoogleSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"client_id": {
			Description: "Client identifier of a registered Google OAuth application.",
			Type:        types.StringType,
			Required:    true,
		},
		"client_secret": {
			Description: "Client secret issued by Google.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
		"hosted_domain": {
			Description: "Restrict users to a Google Apps domain.",
			Type:        types.StringType,
			Optional:    true,
		},
	})
}

func GoogleValidators() []tfsdk.AttributeValidator {
	errSumm := "Invalid Google IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate hosted_domain",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				state := &GoogleIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state)
				if diag.HasError() {
					// No attribute to validate
					return
				}

				if !state.HostedDomain.Unknown && !state.HostedDomain.Null {
					if !common.IsValidDomain(state.HostedDomain.Value) {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a valid Google hosted_domain. Got %v",
								state.HostedDomain.Value),
						)
					}
				}
			},
		},
	}
}

func CreateGoogleIDPBuilder(ctx context.Context, mappingMethod string, state *GoogleIdentityProvider) (*cmv1.GoogleIdentityProviderBuilder, error) {
	builder := cmv1.NewGoogleIdentityProvider()
	builder.ClientID(state.ClientID.Value)
	builder.ClientSecret(state.ClientSecret.Value)

	// Mapping method validation. if mappingMethod != lookup, then hosted-domain is mandatory.
	if mappingMethod != string(cmv1.IdentityProviderMappingMethodLookup) {
		if state.HostedDomain.Unknown || state.HostedDomain.Null {
			return nil, fmt.Errorf("Expected a valid hosted_domain since mapping_method is set to %s", mappingMethod)
		}
	}

	if !state.HostedDomain.Unknown && !state.HostedDomain.Null {
		builder.HostedDomain(state.HostedDomain.Value)
	}

	return builder, nil
}
