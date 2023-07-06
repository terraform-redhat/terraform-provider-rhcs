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
variable "claims" {
  type = object({
    email              = optional(list(string***REMOVED******REMOVED***
    groups             = optional(list(string***REMOVED******REMOVED***
    name               = optional(list(string***REMOVED******REMOVED***
    preferred_username = optional(list(string***REMOVED******REMOVED***
  }***REMOVED***
  default = null
}
variable "extra_scopes" {
  type    = list(string***REMOVED***
  default = null
}
variable "extra_authorize_parameters" {
  type    = list(string***REMOVED***
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