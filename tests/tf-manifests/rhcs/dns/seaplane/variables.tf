# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API base URL"
}

variable "aws_region" {
  type        = string
  description = "AWS region used to sign Seaplane API requests"
}

variable "cluster_arch" {
  type        = string
  description = "DNS-domain cluster architecture. Only HCP is supported by Seaplane."
  default     = "hcp"
}
