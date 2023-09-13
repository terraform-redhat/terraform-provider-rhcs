package proxy

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ProxyResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"http_proxy": schema.StringAttribute{
			Description: "HTTP proxy.",
			Optional:    true,
		},
		"https_proxy": schema.StringAttribute{
			Description: "HTTPS proxy.",
			Optional:    true,
		},
		"no_proxy": schema.StringAttribute{
			Description: "No proxy.",
			Optional:    true,
		},
		"additional_trust_bundle": schema.StringAttribute{
			Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
			Optional:    true,
		},
	}
}
