package identityprovider

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

type GoogleIdentityProvider struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	HostedDomain types.String `tfsdk:"hosted_domain"`
}

var googleSchema = map[string]schema.Attribute{
	"client_id": schema.StringAttribute{
		Description: "Client identifier of a registered Google OAuth application.",
		Required:    true,
	},
	"client_secret": schema.StringAttribute{
		Description: "Client secret issued by Google.",
		Required:    true,
		Sensitive:   true,
	},
	"hosted_domain": schema.StringAttribute{
		Description: "Restrict users to a Google Apps domain.",
		Optional:    true,
		Validators: []validator.String{
			googleHostedDomainValidator(***REMOVED***,
***REMOVED***,
	},
}

func googleHostedDomainValidator(***REMOVED*** validator.String {
	errSumm := "Invalid Google IDP resource configuration"
	return attrvalidators.NewStringValidator("", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
		hostedDomain := req.ConfigValue
		if !hostedDomain.IsNull(***REMOVED*** && !hostedDomain.IsUnknown(***REMOVED*** && !common.IsValidDomain(hostedDomain.ValueString(***REMOVED******REMOVED*** {
			resp.Diagnostics.AddAttributeError(req.Path, errSumm,
				fmt.Sprintf("Expected a valid Google hosted_domain. Got %v",
					hostedDomain.ValueString(***REMOVED******REMOVED******REMOVED***

***REMOVED***
	}***REMOVED***
}

func CreateGoogleIDPBuilder(ctx context.Context, mappingMethod string, state *GoogleIdentityProvider***REMOVED*** (*cmv1.GoogleIdentityProviderBuilder, error***REMOVED*** {
	builder := cmv1.NewGoogleIdentityProvider(***REMOVED***
	builder.ClientID(state.ClientID.ValueString(***REMOVED******REMOVED***
	builder.ClientSecret(state.ClientSecret.ValueString(***REMOVED******REMOVED***

	// Mapping method validation. if mappingMethod != lookup, then hosted-domain is mandatory.
	if mappingMethod != string(cmv1.IdentityProviderMappingMethodLookup***REMOVED*** {
		if state.HostedDomain.IsUnknown(***REMOVED*** || state.HostedDomain.IsNull(***REMOVED*** {
			return nil, fmt.Errorf("Expected a valid hosted_domain since mapping_method is set to %s", mappingMethod***REMOVED***
***REMOVED***
	}

	if !state.HostedDomain.IsUnknown(***REMOVED*** && !state.HostedDomain.IsNull(***REMOVED*** {
		builder.HostedDomain(state.HostedDomain.ValueString(***REMOVED******REMOVED***
	}

	return builder, nil
}
