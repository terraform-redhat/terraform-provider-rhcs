terraform {
  required_providers {
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource "rhcs_default_ingress" "default_ingress" {
  cluster                          = var.cluster
  excluded_namespaces              = var.excluded_namespaces
  route_selectors                  = var.route_selectors
  route_namespace_ownership_policy = var.route_namespace_ownership_policy
  route_wildcard_policy            = var.route_wildcard_policy
  cluster_routes_hostname          = var.cluster_routes_hostname
  cluster_routes_tls_secret_ref    = var.cluster_routes_tls_secret_ref
  id                               = var.id
  load_balancer_type               = var.load_balancer_type


}