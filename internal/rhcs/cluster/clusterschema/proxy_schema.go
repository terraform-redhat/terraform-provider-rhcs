package clusterschema

import (
	"fmt"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/rosa/pkg/helper"
)

func ProxyFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"http_proxy": {
			Description: "http proxy",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"https_proxy": {
			Description: "https proxy",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"no_proxy": {
			Description: "no proxy",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"additional_trust_bundle": {
			Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

type Proxy struct {
	HttpProxy             *string `tfsdk:"http_proxy"`
	HttpsProxy            *string `tfsdk:"https_proxy"`
	NoProxy               *string `tfsdk:"no_proxy"`
	AdditionalTrustBundle *string `tfsdk:"additional_trust_bundle"`
}

func ExpandProxyFromInterface(i interface{}) *Proxy {
	l := i.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	proxyMap := l[0].(map[string]interface{})
	return &Proxy{
		HttpProxy:             common.GetOptionalStringFromMapString(proxyMap, "http_proxy"),
		HttpsProxy:            common.GetOptionalStringFromMapString(proxyMap, "https_proxy"),
		NoProxy:               common.GetOptionalStringFromMapString(proxyMap, "no_proxy"),
		AdditionalTrustBundle: common.GetOptionalStringFromMapString(proxyMap, "additional_trust_bundle"),
	}
}

func ExpandProxyFromResourceData(resourceData *schema.ResourceData) *Proxy {
	list, ok := resourceData.GetOk("proxy")
	if !ok {
		return nil
	}

	return ExpandProxyFromInterface(list)
}

func FlatProxy(object *cmv1.Cluster, resourceData *schema.ResourceData) []interface{} {
	result := make(map[string]interface{})

	if proxy, ok := object.GetProxy(); ok {

		if httpProxy, ok := proxy.GetHTTPProxy(); ok {
			result["http_proxy"] = httpProxy
		}

		if httpsProxy, ok := proxy.GetHTTPSProxy(); ok {
			result["https_proxy"] = httpsProxy
		}

		if noProxy, ok := proxy.GetNoProxy(); ok {
			result["no_proxy"] = noProxy
		}
	}

	originalTrustedBundle := getOriginalTrustedBundleValue(resourceData)
	if originalTrustedBundle != "" {
		result["additional_trust_bundle"] = originalTrustedBundle
	} else if trustBundle, ok := object.GetAdditionalTrustBundle(); ok {
		result["additional_trust_bundle"] = trustBundle
	}

	return []interface{}{result}
}

func getOriginalTrustedBundleValue(resourceData *schema.ResourceData) string {
	proxy := ExpandProxyFromResourceData(resourceData)
	if proxy == nil {
		return ""
	}

	if proxy.AdditionalTrustBundle != nil {
		return *proxy.AdditionalTrustBundle
	}

	return ""
}

func ProxyValidators(i interface{}) error {
	proxy := ExpandProxyFromInterface(i)
	if proxy == nil {
		return nil
	}
	httpsProxy := ""
	if proxy.HttpsProxy != nil {
		httpsProxy = *proxy.HttpsProxy
	}

	httpProxy := ""
	if proxy.HttpProxy != nil {
		httpProxy = *proxy.HttpProxy
	}

	additionalTrustBundle := ""
	if proxy.AdditionalTrustBundle != nil {
		additionalTrustBundle = *proxy.AdditionalTrustBundle
	}

	var noProxySlice []string
	if proxy.NoProxy != nil {
		noProxySlice = helper.HandleEmptyStringOnSlice(strings.Split(*proxy.NoProxy, ","))
	}

	if httpProxy == "" && httpsProxy == "" && noProxySlice != nil && len(noProxySlice) > 0 {
		return fmt.Errorf("Expected at least one of the following: http-proxy, https-proxy")

	}

	if httpProxy == "" && httpsProxy == "" && additionalTrustBundle == "" {
		return fmt.Errorf("Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle")
	}

	return nil
}
