package proxy

import (
	dsschemadsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const (
	resetMessage = "To reset please provide '' (empty string)"
)

func ProxyResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"http_proxy": schema.StringAttribute{
			Description: "HTTP proxy. " + resetMessage,
			Optional:    true,
		},
		"https_proxy": schema.StringAttribute{
			Description: "HTTPS proxy. " + resetMessage,
			Optional:    true,
		},
		"no_proxy": schema.StringAttribute{
			Description: "No proxy. " + resetMessage,
			Optional:    true,
		},
		"additional_trust_bundle": schema.StringAttribute{
			Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store. " + resetMessage,
			Optional:    true,
		},
	}
}

func ProxyDatasource() map[string]dsschemadsschema.Attribute {
	return map[string]dsschemadsschema.Attribute{
		"http_proxy": schema.StringAttribute{
			Description: "HTTP proxy.",
			Computed:    true,
		},
		"https_proxy": schema.StringAttribute{
			Description: "HTTPS proxy.",
			Computed:    true,
		},
		"no_proxy": schema.StringAttribute{
			Description: "No proxy.",
			Computed:    true,
		},
		"additional_trust_bundle": schema.StringAttribute{
			Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
			Computed:    true,
		},
	}
}
