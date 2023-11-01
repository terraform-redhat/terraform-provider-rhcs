variable "token" {
  type        = string
  description = "RHCS token - You can get it here: https://console.redhat.com/openshift/token"
}

variable "cluster_id" {
  type        = string
  description = "The OCP cluster ID"
}

variable "url" {
  type        = string
  description = "Provide RHCS environment by setting a value to url"
  default     = "https://api.stage.openshift.com"
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
  type        = list
  # description = "List of GitHub organizations"
  default = ["aaa","ddd"]
}
variable "name" {
    type= string
    default = "Github"
}