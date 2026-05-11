# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "cluster_id" {
  type    = string
  default = null
}

variable "timeout_in_min" {
  type    = number
  default = 60
}