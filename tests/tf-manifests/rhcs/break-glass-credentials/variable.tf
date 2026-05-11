# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "cluster" {
  type = string
}

variable "expiration_duration" {
  type    = string
  default = null
}

variable "username" {
  type    = string
  default = null
}
