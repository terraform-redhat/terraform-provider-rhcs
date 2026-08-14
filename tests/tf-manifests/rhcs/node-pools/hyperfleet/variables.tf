# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "hyperfleet_url" {
  type        = string
  description = "Hyperfleet Platform API v2 base URL"
}

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "cluster_id" {
  type        = string
  description = "UID of the parent rhcs_cluster_hyperfleet (its `id` output)"
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
  description = "Fixed replica count (mutually exclusive with autoscaling)"
  default     = null
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type (e.g. m5.xlarge)"
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
