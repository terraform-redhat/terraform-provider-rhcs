# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "oidc_config_id" {
  value = rhcs_oidcconfig_seaplane.oidc.id
}

output "name" {
  value = rhcs_oidcconfig_seaplane.oidc.name
}

output "issuer_url" {
  value = rhcs_oidcconfig_seaplane.oidc.issuer_url
}

output "oidc_config_phase" {
  value = rhcs_oidcconfig_seaplane.oidc.phase
}

output "thumbprint" {
  value = rhcs_oidcconfig_seaplane.oidc.thumbprint
}
