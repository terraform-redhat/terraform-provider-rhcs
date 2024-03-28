package proxy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type Proxy struct {
	HttpProxy             types.String `tfsdk:"http_proxy"`
	HttpsProxy            types.String `tfsdk:"https_proxy"`
	NoProxy               types.String `tfsdk:"no_proxy"`
	AdditionalTrustBundle types.String `tfsdk:"additional_trust_bundle"`
}

func BuildProxy(state *Proxy, builder *cmv1.ClusterBuilder) (*cmv1.ClusterBuilder, error) {
	if state != nil {
		proxy := cmv1.NewProxy()
		proxyIsEmpty := true

		if common.HasValue(state.HttpProxy) {
			proxy.HTTPProxy(state.HttpProxy.ValueString())
			proxyIsEmpty = false
		}
		if common.HasValue(state.HttpsProxy) {
			proxy.HTTPSProxy(state.HttpsProxy.ValueString())
			proxyIsEmpty = false
		}
		if common.HasValue(state.NoProxy) {
			proxy.NoProxy(state.NoProxy.ValueString())
			proxyIsEmpty = false
		}
		if !proxyIsEmpty {
			builder.Proxy(proxy)
		}

		if common.HasValue(state.AdditionalTrustBundle) {
			builder.AdditionalTrustBundle(state.AdditionalTrustBundle.ValueString())
		}

	}

	return builder, nil
}
