variable "account_role_prefix" {
  type = string
}

variable "account_role_properties" {
  description = "Account IAM role properties"
  type = object({
    role_name      = string
    role_type      = string
    principal      = string
    policy_details = string
  })
}

variable "instance_account_role_properties" {
  description = "Account IAM role properties"
  type = object({
    role_name      = string
    role_type      = string
    policy_details = string
  })
}

variable "rosa_openshift_version" {
  type    = string
  default = "4.12"
}

variable "account_id" {
  type = string
  default = "local"
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
