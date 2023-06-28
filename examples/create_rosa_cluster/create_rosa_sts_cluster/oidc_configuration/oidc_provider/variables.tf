variable "token" {
  type      = string
  sensitive = true
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

variable "managed" {
  description = "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted) OIDC Configuration"
  type        = bool
}

variable "installer_role_arn" {
  description = "STS Role ARN with get secrets permission, relevant only for unmanaged OIDC config"
  type        = string
  default     = null
}

variable "operator_role_prefix" {
  type = string
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "cloud_region" {
  type    = string
  default = "us-east-2"
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}
