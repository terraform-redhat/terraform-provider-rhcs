variable "token" {
  type      = string
  sensitive = true
}
variable "gateway" {
  type    = string
  default = "https://api.stage.openshift.com"
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
variable "ca" {
  type    = string
  default = null
}
variable "hostname" {
  type    = string
  default = null
}
variable "organizations" {
  type    = list(string)
  default = null
}
variable "teams" {
  type    = list(string)
  default = null
}
variable "url" {
  type    = string
  default = null
}
