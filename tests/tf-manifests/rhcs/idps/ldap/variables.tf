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
variable "bind_dn" {
  type    = string
  default = null
}

variable "bind_password" {
  type    = string
  default = null
}

variable "insecure" {
  type    = bool
  default = false
}

variable "attributes" {
  type = object({
    email              = optional(list(string***REMOVED******REMOVED***
    id                 = optional(list(string***REMOVED******REMOVED***
    name               = optional(list(string***REMOVED******REMOVED***
    preferred_username = optional(list(string***REMOVED******REMOVED***
  }***REMOVED***
  default = null
}

variable "ca" {
  type    = string
  default = null
}
variable "url" {
  type    = string
  default = null
}