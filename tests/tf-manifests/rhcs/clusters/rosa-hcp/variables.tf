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
  default = false
}

variable "ec2_metadata_http_tokens" {
  type    = string
  default = null
}

variable "host_prefix" {
  type    = number
  default = 24
}

variable "etcd_kms_key_arn" {
  type    = string
  default = null
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

variable "fips" {
  type    = bool
  default = false
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

variable "aws_account_id" {
  type    = string
  default = null
}

variable "aws_billing_account_id" {
  type    = string
  default = null
}

variable "installer_role" {
  type    = string
  default = null
}

variable "support_role" {
  type    = string
  default = null
}

variable "worker_role" {
  type    = string
  default = null
}

variable "include_creator_property" {
  type    = bool
  default = true
}

variable "wait_for_cluster" {
  type    = bool
  default = true
}

variable "disable_cluster_waiter" {
  type    = bool
  default = false
}

variable "additional_infra_security_groups" {
  type    = list(string)
  default = null
}

variable "additional_control_plane_security_groups" {
  type    = list(string)
  default = null
}

variable "disable_waiting_in_destroy" {
  type    = bool
  default = false
}

variable "full_resources" {
  type    = bool
  default = false
}

variable "registry_config" {
  type = object({
    additional_trusted_ca = optional(map(string))
    allowed_registries_for_import = optional(
      list(
        object(
          {
            domain_name = optional(string)
            insecure    = optional(bool)
          }
        )
      )
    )
    platform_allowlist_id = optional(string)
    registry_sources = optional(
      object(
        {
          allowed_registries  = optional(list(string))
          blocked_registries  = optional(list(string))
          insecure_registries = optional(list(string))
        }
      )
    )
  })
  default = null
}

variable "worker_disk_size" {
  type    = number
  default = null
}

variable "additional_compute_security_groups" {
  type    = list(string)
  default = null
}
