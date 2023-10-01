variable "token" {
  type      = string
  sensitive = true
}

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
  default = "rhcs-tf"
}

variable "hosted_cp" {
  type    = bool
  default = false
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

variable "autoscaling" {
  type = object({
    autoscaling_enabled = bool
    min_replicas        = optional(number)
    max_replicas        = optional(number)
  })
  default = {
    autoscaling_enabled = false
  }
}

variable "aws_http_tokens_state" {
  type    = string
  default = null
}

variable "private_link" {
  type    = bool
  default = false
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

variable "default_mp_labels" {
  type    = map(string)
  default = null
}

variable "disable_scp_checks" {
  type    = bool
  default = false
}
variable "disable_workload_monitoring" {
  type    = bool
  default = false
}
variable "etcd_encryption" {
  type    = bool
  default = false
}

variable "fips" {
  type    = bool
  default = false
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

variable "properties" {
  type    = map(string)
  default = null
}

variable "proxy" {
  type = object({
    http_proxy              = string
    https_proxy             = string
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
  type        = string
  default     = null

}
