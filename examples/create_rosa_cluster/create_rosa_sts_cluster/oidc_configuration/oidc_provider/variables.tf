variable "token" {
  type      = string
  sensitive = true
}

variable "url" {
  type    = string
  default = "https://api.openshift.com"
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
