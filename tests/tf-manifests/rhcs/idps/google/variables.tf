variable "token" {
  type      = string
  sensitive = true
}
variable "url" {
  type        = string
  description = "Provide RHCS environment by setting a value to url"
  default     = "https://api.stage.openshift.com"
}

// Shared by all of the IDPs
variable "cluster_id" {
  type = string
}
variable "name" {
  type = string
}
variable "mapping_method" {
  type    = string
  default = "claim"
}
variable "client_id" {
  type    = string
  default = null
}
variable "client_secret" {
  type    = string
  default = null
}
variable "hosted_domain" {
  type    = string
  default = null
}