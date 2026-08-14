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

variable "cluster_name" {
  type        = string
  description = "Human-readable cluster name"
}

variable "operator_roles_prefix" {
  type        = string
  description = "Prefix for the seven operator IAM roles"
}

variable "subnet_id" {
  type        = string
  description = "Worker-node subnet ID"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID containing the worker subnet"
}

variable "availability_zone" {
  type        = string
  description = "Availability zone of the worker subnet"
}

variable "expiration_timestamp" {
  type        = string
  description = "Optional RFC3339 expiration timestamp for automatic cluster deletion"
  default     = null
}
