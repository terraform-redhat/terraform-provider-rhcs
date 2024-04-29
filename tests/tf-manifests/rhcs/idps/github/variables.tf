variable "cluster_id" {
  type        = string
  description = "The OCP cluster ID"
}

# IDP Variables
variable "client_id" {
  type        = string
  description = "GitHub client id"
}
variable "client_secret" {
  type        = string
  description = "GitHub client secret"
}
variable "mapping_method" {
  type    = string
  default = "claim"
}
variable "organizations" {
  type = list(any)
  # description = "List of GitHub organizations"
  default = ["aaa", "ddd"]
}
variable "name" {
  type = string
}