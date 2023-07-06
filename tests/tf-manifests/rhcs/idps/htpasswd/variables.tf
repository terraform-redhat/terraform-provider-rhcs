variable "token" {
  type      = string
  sensitive = true
}
variable "gateway" {
  type    = string
  default = "https://api.openshift.com" // default to production env. Once run on another gateway, set it 
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
variable "username" {
  type    = string
  default = null
}
variable "password" {
  type    = string
  default = null
}