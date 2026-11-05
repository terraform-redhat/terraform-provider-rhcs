# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API base URL"
}

variable "operator_roles_prefix" {
  type        = string
  description = "Operator roles prefix used to name the IAM roles"
}

variable "oidc_issuer_url" {
  type        = string
  description = "OIDC issuer URL the operator role trust policies are bound to"
}
