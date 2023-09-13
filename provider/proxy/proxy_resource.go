package proxy

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
***REMOVED***

func ProxyResource(***REMOVED*** map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"http_proxy": schema.StringAttribute{
			Description: "HTTP proxy.",
			Optional:    true,
***REMOVED***,
		"https_proxy": schema.StringAttribute{
			Description: "HTTPS proxy.",
			Optional:    true,
***REMOVED***,
		"no_proxy": schema.StringAttribute{
			Description: "No proxy.",
			Optional:    true,
***REMOVED***,
		"additional_trust_bundle": schema.StringAttribute{
			Description: "A string containing a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store.",
			Optional:    true,
***REMOVED***,
	}
}
