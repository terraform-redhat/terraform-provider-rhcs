package proxy

import "github.com/hashicorp/terraform-plugin-framework/types"

type Proxy struct {
	HttpProxy             types.String `tfsdk:"http_proxy"`
	HttpsProxy            types.String `tfsdk:"https_proxy"`
	NoProxy               types.String `tfsdk:"no_proxy"`
	AdditionalTrustBundle types.String `tfsdk:"additional_trust_bundle"`
}
