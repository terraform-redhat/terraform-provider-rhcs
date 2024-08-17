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

variable "ec2_metadata_http_tokens" {
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

variable "admin_credentials" {
  description = "Admin user and password"
  type        = map(string)
  default     = null

}

variable "worker_disk_size" {
  type    = number
  default = null

}

variable "additional_compute_security_groups" {
  type    = list(string)
  default = null

}

variable "additional_infra_security_groups" {
  type    = list(string)
  default = null

}

variable "additional_control_plane_security_groups" {
  type    = list(string)
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

variable "base_dns_domain" {
  type    = string
  default = null
}

variable "domain_prefix" {
  type    = string
  default = null
}

variable "private_hosted_zone" {
  type = object({
    id       = string
    role_arn = string
  })
  default = null
}

variable "wait_for_cluster" {
  type    = bool
  default = true
}

variable "disable_cluster_waiter" {
  type    = bool
  default = false
}

variable "disable_waiting_in_destroy" {
  type    = bool
  default = false
}

variable "full_resources" {
  type    = bool
  default = false
}
