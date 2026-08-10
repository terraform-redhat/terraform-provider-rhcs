# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "operator_roles_prefix" {
  type        = string
  description = "Prefix for the seven operator IAM roles and the worker role"
}

variable "oidc_issuer_url" {
  type        = string
  description = "OIDC issuer URL returned by the hyperfleet cluster (e.g. https://s3.amazonaws.com/...)"
}
