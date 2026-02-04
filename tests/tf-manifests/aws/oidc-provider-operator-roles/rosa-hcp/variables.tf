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

variable "oidc_prefix" {
  type        = string
  description = "Optional prefix for the OIDC resources (if you're using managed policies). Maximum 16 characters, must match pattern: ^[a-z][a-z0-9\\-]+[a-z0-9]$"
  default     = null

  validation {
    condition     = var.oidc_prefix == null || length(var.oidc_prefix) <= 16
    error_message = "The oidc_prefix must be maximum 16 characters"
  }

  validation {
    condition     = var.oidc_prefix == null || can(regex("^[a-z][a-z0-9\\-]+[a-z0-9]$", var.oidc_prefix))
    error_message = "The oidc_prefix must start with a lowercase letter, contain only lowercase letters/numbers/hyphens, and end with a lowercase letter or number. Pattern: ^[a-z][a-z0-9\\-]+[a-z0-9]$"
  }
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
