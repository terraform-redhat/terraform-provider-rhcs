variable "ocm_environment" {
  type    = string
  default = "production"
}

variable "openshift_version" {
  type    = string
  default = "4.13"
}

variable "account_role_prefix" {
  type = string
}

variable "token" {
  type = string
}

variable "url" {
  type = string
}

variable "account_role_path" {
  description = "(Optional***REMOVED*** Path to the account role."
  type        = string
  default     = null
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string***REMOVED***
  default     = null
}

