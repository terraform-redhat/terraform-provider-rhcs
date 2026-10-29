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

variable "name" {
  type        = string
  description = "Customer-selected DNS reservation name"
}
