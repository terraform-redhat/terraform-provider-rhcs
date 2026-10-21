# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API base URL"
}

variable "name" {
  type        = string
  description = "Client-selected name of the OIDC configuration"
}

variable "type" {
  type        = string
  description = "Type of the OIDC configuration (managed or unmanaged)"
  default     = "managed"
}
