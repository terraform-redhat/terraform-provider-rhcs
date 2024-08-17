variable "cluster" {
  type = string
}

variable "cluster_routes_hostname" {
  type    = string
  default = null
}

variable "cluster_routes_tls_secret_ref" {
  type    = string
  default = null
}
variable "excluded_namespaces" {
  type    = list(string)
  default = null
}
variable "id" {
  type    = string
  default = null
}

variable "load_balancer_type" {
  type    = string
  default = null
}
variable "route_namespace_ownership_policy" {
  type    = string
  default = null
}
variable "route_selectors" {
  type    = map(string)
  default = null
}
variable "route_wildcard_policy" {
  type    = string
  default = null
}

variable "component_routes" {
  type = map(object({
    hostname       = string
    tls_secret_ref = string
  }))
  default = null
}
