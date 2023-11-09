variable "ocm_environment" {
  type    = string
  default = "production"
}

variable "openshift_version" {
  type = string
  default = ""
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "token" {
  type = string
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}
