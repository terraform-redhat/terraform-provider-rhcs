package classic

import "github.com/hashicorp/terraform-plugin-framework/types"

type DefaultIngress struct {
	Cluster                   types.String `tfsdk:"cluster"`
	RouteSelectors            types.Map    `tfsdk:"route_selectors"`
	ExcludedNamespaces        types.List   `tfsdk:"excluded_namespaces"`
	WildcardPolicy            types.String `tfsdk:"route_wildcard_policy"`
	NamespaceOwnershipPolicy  types.String `tfsdk:"route_namespace_ownership_policy"`
	Id                        types.String `tfsdk:"id"`
	ClusterRoutesHostname     types.String `tfsdk:"cluster_routes_hostname"`
	ClusterRoutesTlsSecretRef types.String `tfsdk:"cluster_routes_tls_secret_ref"`
	LoadBalancerType          types.String `tfsdk:"load_balancer_type"`
}
