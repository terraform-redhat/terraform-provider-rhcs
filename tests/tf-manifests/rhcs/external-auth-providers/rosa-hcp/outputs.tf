# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "claim_mapping_groups_claim" {
  description = "Configured groups claim mapping."
  value       = rhcs_external_auth_provider.main.claim_mapping_groups_claim
}

output "claim_mapping_username_claim" {
  description = "Configured username claim mapping."
  value       = rhcs_external_auth_provider.main.claim_mapping_username_claim
}

output "claim_validation_rule" {
  description = "Configured token claim validation rules."
  value       = rhcs_external_auth_provider.main.claim_validation_rule
}

output "cluster" {
  description = "ID of the cluster that owns the provider."
  value       = rhcs_external_auth_provider.main.cluster
}

output "id" {
  description = "ID of the external authentication provider."
  value       = rhcs_external_auth_provider.main.id
}

output "issuer_audiences" {
  description = "Audiences accepted by the external token issuer."
  value       = rhcs_external_auth_provider.main.issuer_audiences
}

output "issuer_url" {
  description = "HTTPS URL of the external token issuer."
  value       = rhcs_external_auth_provider.main.issuer_url
}

output "name" {
  description = "Name of the external authentication provider."
  value       = rhcs_external_auth_provider.main.name
}
