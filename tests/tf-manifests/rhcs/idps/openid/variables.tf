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
variable "claims" {
  type = object({
    email              = optional(list(string))
    groups             = optional(list(string))
    name               = optional(list(string))
    preferred_username = optional(list(string))
  })
  default = null
}
variable "extra_scopes" {
  type    = list(string)
  default = null
}
variable "extra_authorize_parameters" {
  type    = list(string)
  default = null
}
variable "issuer" {
  type    = string
  default = null
}

variable "ca" {
  type    = string
  default = null
}

variable "client_id" {
  type    = string
  default = null
}

variable "client_secret" {
  type    = string
  default = null
}