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

variable "image_type" {
  type        = string
  description = "Operating system image family (Linux or Windows)"
  default     = null
}

variable "node_drain_grace_period" {
  type        = string
  description = "Grace period before a node is forcibly drained during upgrade (e.g. 5m, 1h)"
  default     = null
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type (e.g. m5.xlarge)"
}

variable "placement_market_type" {
  type        = string
  description = "EC2 purchasing model, such as OnDemand or Spot"
  default     = null
}

variable "max_price" {
  type        = string
  description = "Maximum hourly price in USD for Spot instances"
  default     = null
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
