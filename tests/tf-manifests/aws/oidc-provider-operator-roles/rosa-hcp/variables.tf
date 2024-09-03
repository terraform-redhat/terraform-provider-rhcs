variable "operator_role_prefix" {
  type    = string
  default = ""
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "oidc_config" {
  type    = string
  default = ""
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "List of AWS resource tags to apply."
}

variable "permissions_boundary" {
  description = "The ARN of the policy that is used to set the permissions boundary for the IAM roles in STS clusters."
  type        = string
  default     = ""
}
