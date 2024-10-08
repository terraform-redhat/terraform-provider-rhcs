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
variable "htpasswd_users" {
  type = list(object({
    username = string
    password = string
  }))
  description = "htpasswd user list"
  sensitive   = true
}