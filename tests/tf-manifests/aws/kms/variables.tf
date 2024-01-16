variable "aws_region" {
  type = string
}

variable "kms_name" {
  type = string
}


variable "tag_key" {
  type    = string
  default = ""
}

variable "tag_value" {
  type    = string
  default = ""
}

variable "tag_description" {
  type    = string
  default = ""
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}