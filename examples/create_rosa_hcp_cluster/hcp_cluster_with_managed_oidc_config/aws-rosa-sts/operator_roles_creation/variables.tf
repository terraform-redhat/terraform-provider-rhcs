variable "cluster_id" {
  description = "cluster ID"
  type        = string
}

variable "rh_oidc_provider_url" {
  description = "oidc provider url"
  type        = string
  default     = "rh-oidc.s3.us-east-1.amazonaws.com"
}

variable "operator_roles_properties" {
  description = "List of ROSA Operator IAM Roles"
  type = list(object({
    role_name          = string
    policy_name        = string
    service_accounts   = list(string)
    operator_name      = string
    operator_namespace = string
  }))
  validation {
    condition     = length(var.operator_roles_properties) == 8
    error_message = "The list of operator roles should contains 8 elements."
  }
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}

variable "permissions_boundary" {
  description = "The ARN of the policy that is used to set the permissions boundary for the IAM roles in STS clusters."
  type        = string
  default     = ""
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}
