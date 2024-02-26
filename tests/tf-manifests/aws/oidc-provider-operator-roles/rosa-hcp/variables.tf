variable "operator_role_prefix" {
  type = string
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "oidc_config" {
  type    = string
  default = ""
}

variable "installer_role_arn" {
  type        = string
  default     = null
  description = "The Amazon Resource Name (ARN) associated with the AWS IAM role used by the ROSA installer. Applicable exclusively to unmanaged OIDC; otherwise, leave empty."
}

variable "aws_region" {
  type    = string
  default = "us-east-2"
}

variable "rhcs_environment" {
  type    = string
  default = "staging"
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

variable "cluster_id" {
  description = "cluster ID"
  type        = string
  default     = ""
}