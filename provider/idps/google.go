package idps

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
***REMOVED***

type GoogleIdentityProvider struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	HostedDomain types.String `tfsdk:"hosted_domain"`
}

func GoogleSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"client_id": {
			Description: "Client identifier of a registered Google OAuth application.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"client_secret": {
			Description: "Client secret issued by Google.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
		"hosted_domain": {
			Description: "Restrict users to a Google Apps domain.",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
	}***REMOVED***
}

func GoogleValidators(***REMOVED*** []tfsdk.AttributeValidator {
	errSumm := "Invalid Google IDP resource configuration"
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate hosted_domain",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
				state := &GoogleIdentityProvider{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, state***REMOVED***
				if diag.HasError(***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				if !state.HostedDomain.Unknown && !state.HostedDomain.Null {
					if !common.IsValidDomain(state.HostedDomain.Value***REMOVED*** {
						resp.Diagnostics.AddError(errSumm,
							fmt.Sprintf("Expected a valid Google hosted_domain. Got %v",
								state.HostedDomain.Value***REMOVED***,
						***REMOVED***
			***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}

func CreateGoogleIDPBuilder(ctx context.Context, mappingMethod string, state *GoogleIdentityProvider***REMOVED*** (*cmv1.GoogleIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewGoogleIdentityProvider(***REMOVED***
	builder.ClientID(state.ClientID.Value***REMOVED***
	builder.ClientSecret(state.ClientSecret.Value***REMOVED***

	// Mapping method validation. if mappingMethod != lookup, then hosted-domain is mandatory.
	if mappingMethod != string(cmv1.IdentityProviderMappingMethodLookup***REMOVED*** {
		if state.HostedDomain.Unknown || state.HostedDomain.Null {
			return nil, fmt.Errorf("Expected a valid hosted_domain since mapping_method is set to %s", mappingMethod***REMOVED***
***REMOVED***
	}

	if !state.HostedDomain.Unknown && !state.HostedDomain.Null {
		builder.HostedDomain(state.HostedDomain.Value***REMOVED***
	}

	return builder, nil
}
