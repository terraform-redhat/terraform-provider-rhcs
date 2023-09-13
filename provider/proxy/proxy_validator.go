package proxy

***REMOVED***
	"context"
***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/openshift/rosa/pkg/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type proxyValidator struct {
}

// Description describes the validation in plain text formatting.
func (v proxyValidator***REMOVED*** Description(_ context.Context***REMOVED*** string {
	return fmt.Sprintf("proxy map should not include an hard coded OCM proxy"***REMOVED***
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v proxyValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.Description(ctx***REMOVED***
}

// Validate performs the validation.
func (v proxyValidator***REMOVED*** ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse***REMOVED*** {
	if request.ConfigValue.IsNull(***REMOVED*** || request.ConfigValue.IsUnknown(***REMOVED*** {
		return
	}

	proxy := Proxy{}
	d := request.ConfigValue.As(ctx, &proxy, basetypes.ObjectAsOptions{}***REMOVED***
	if d.HasError(***REMOVED*** {
		// No attribute to validate
		return
	}
	errSum := "Invalid proxy's attribute assignment"
	httpsProxy := ""
	httpProxy := ""
	additionalTrustBundle := ""
	var noProxySlice []string

	if !common.IsStringAttributeEmpty(proxy.HttpProxy***REMOVED*** {
		httpProxy = proxy.HttpProxy.ValueString(***REMOVED***
	}
	if !common.IsStringAttributeEmpty(proxy.HttpsProxy***REMOVED*** {
		httpsProxy = proxy.HttpsProxy.ValueString(***REMOVED***
	}
	if !common.IsStringAttributeEmpty(proxy.NoProxy***REMOVED*** {
		noProxySlice = helper.HandleEmptyStringOnSlice(strings.Split(proxy.NoProxy.ValueString(***REMOVED***, ","***REMOVED******REMOVED***
	}

	if !common.IsStringAttributeEmpty(proxy.AdditionalTrustBundle***REMOVED*** {
		additionalTrustBundle = proxy.AdditionalTrustBundle.ValueString(***REMOVED***
	}

	if httpProxy == "" && httpsProxy == "" && noProxySlice != nil && len(noProxySlice***REMOVED*** > 0 {
		response.Diagnostics.AddError(errSum, "Expected at least one of the following: http-proxy, https-proxy"***REMOVED***
		return
	}
	if httpProxy == "" && httpsProxy == "" && additionalTrustBundle == "" {
		response.Diagnostics.AddError(errSum, "Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle"***REMOVED***
		return
	}
}

func ProxyValidator(***REMOVED*** validator.Object {
	return proxyValidator{}
}
