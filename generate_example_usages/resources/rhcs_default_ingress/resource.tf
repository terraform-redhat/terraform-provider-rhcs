resource "rhcs_default_ingress" "default_ingress" {
  cluster          = "cluster-id-123"
  excluded_namespaces = ["example_ns"]
  route_wildcard_policy = "WildcardsAllowed"
  route_namespace_ownership_policy = "InterNamespaceAllowed"
  load_balancer_type = "nlb"
}
