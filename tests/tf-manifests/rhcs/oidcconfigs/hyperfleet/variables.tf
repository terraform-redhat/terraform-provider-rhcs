# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "hyperfleet_url" {
  type        = string
  description = "Hyperfleet Platform API v2 base URL"
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
