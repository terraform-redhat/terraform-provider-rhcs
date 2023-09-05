variable "account_role_prefix" {
  type = string
}

variable "operator_role_policy_properties" {
  description = "Account IAM role properties"
  type = object({
    policy_name    = string
    policy_details = string
    namespace      = string
    operator_name  = string
  })
}

variable "rosa_openshift_version" {
  type    = string
  default = "4.12"
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}
