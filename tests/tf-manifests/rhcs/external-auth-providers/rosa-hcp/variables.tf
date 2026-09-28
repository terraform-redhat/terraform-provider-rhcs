# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "claim_mapping_groups_claim" {
  description = "Token claim used for group mappings."
  type        = string
}

variable "claim_mapping_username_claim" {
  description = "Token claim used for username mappings."
  type        = string
}

variable "claim_validation_rule" {
  description = "Token claim validation rules in claim:required_value form."
  type        = list(string)
}

variable "cluster" {
  description = "ID of the ready ROSA HCP cluster."
  type        = string
}

variable "issuer_audiences" {
  description = "Audiences accepted by the external token issuer."
  type        = set(string)
}

variable "issuer_url" {
  description = "HTTPS URL of the external token issuer."
  type        = string
}

variable "name" {
  description = "Name of the external authentication provider."
  type        = string
}
