
variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "operator_role_prefix" {
  type    = string
  default = ""
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "cluster_name" {
  type    = string
  default = "rhcs-tf-hcp"
}

variable "aws_availability_zones" {
  type    = list(string)
  default = null
}

variable "replicas" {
  type    = number
  default = 3
}

variable "openshift_version" {
  type    = string
  default = null
}

variable "channel_group" {
  type    = string
  default = "stable"
}

variable "rhcs_environment" {
  type    = string
  default = "staging"
}

variable "product" {
  type    = string
  default = "rosa"
}

variable "private" {
  type    = bool
  default = false
}

variable "aws_subnet_ids" {
  type    = list(string)
  default = null
}

variable "compute_machine_type" {
  type    = string
  default = null
}

variable "etcd_encryption" {
  type    = bool
  default = null # Temporary. Should be false but due to https://issues.redhat.com/browse/OCM-6593 it is set to null by default
}

variable "host_prefix" {
  type    = number
  default = 24
}

variable "kms_key_arn" {
  type    = string
  default = null
}

variable "machine_cidr" {
  type    = string
  default = null
}

variable "service_cidr" {
  type    = string
  default = null
}

variable "pod_cidr" {
  type    = string
  default = null
}

variable "custom_properties" {
  type    = map(string)
  default = null
}

variable "proxy" {
  type = object({
    http_proxy              = optional(string)
    https_proxy             = optional(string)
    additional_trust_bundle = optional(string)
    no_proxy                = optional(string)
  })
  default = null
}

variable "tags" {
  type    = map(string)
  default = null
}

variable "multi_az" {
  type    = bool
  default = false
}

variable "aws_region" {
  type        = string
  description = "The region to create the ROSA cluster in"
}

variable "oidc_config_id" {
  type    = string
  default = null
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}

variable "upgrade_acknowledgements_for" {
  type    = string
  default = null
}
