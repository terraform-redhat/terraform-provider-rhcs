# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "name_prefix" {
  type        = string
  description = "Prefix applied to all resource names (typically the cluster name)"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR block for the VPC"
  default     = "10.0.0.0/16"
}

variable "availability_zone" {
  type        = string
  description = "Availability zone for the private and public subnets"
}
