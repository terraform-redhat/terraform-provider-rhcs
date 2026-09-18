# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "oidc_config_id" {
  value = rhcs_oidcconfig_hyperfleet.oidc.id
}

output "name" {
  value = rhcs_oidcconfig_hyperfleet.oidc.name
}

output "issuer_url" {
  value = rhcs_oidcconfig_hyperfleet.oidc.issuer_url
}

output "oidc_config_phase" {
  value = rhcs_oidcconfig_hyperfleet.oidc.phase
}

output "thumbprint" {
  value = rhcs_oidcconfig_hyperfleet.oidc.thumbprint
}
