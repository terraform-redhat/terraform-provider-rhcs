
variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "token" {
  type = string
}

variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "operator_role_prefix" {
  type    = string
  default = ""
}

variable "oidc_config" {
  type    = string
  default = ""
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