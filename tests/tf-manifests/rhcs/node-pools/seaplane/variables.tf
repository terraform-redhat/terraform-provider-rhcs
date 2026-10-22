# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API base URL"
}

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "cluster_id" {
  type        = string
  description = "UID of the parent rhcs_cluster_seaplane (its `id` output)"
}

variable "cluster_name" {
  type        = string
  description = "Name of the parent rhcs_cluster_seaplane"
}

variable "name" {
  type        = string
  description = "Node pool name"
}

variable "subnet_id" {
  type        = string
  description = "AWS subnet ID for node instances"
}

variable "auto_repair" {
  type        = bool
  description = "Enable auto-repair of unhealthy nodes"
  default     = true
}

variable "replicas" {
  type        = number
  description = "Fixed replica count"
  default     = null
}

variable "autoscaling" {
  type = object({
    enabled      = bool
    min_replicas = optional(number)
    max_replicas = optional(number)
  })
  description = "Optional NodePool autoscaling configuration"
  default     = null
}

variable "max_unavailable" {
  type        = string
  description = "Largest number of nodes that may be unavailable during an upgrade, as an absolute count"
  default     = null
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type (e.g. m5.xlarge)"
}

variable "ec2_instance_metadata_http_tokens" {
  type        = string
  description = "EC2 instance metadata HTTP token requirement"
  default     = null
}

variable "disk_size" {
  type        = number
  description = "Root volume size in GiB"
  default     = null
}

variable "tags" {
  type        = map(string)
  description = "Additional AWS resource tags for node instances"
  default     = null
}

variable "labels" {
  type        = map(string)
  description = "Node labels"
  default     = null
}
