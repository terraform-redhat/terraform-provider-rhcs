# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "idp_google_id" {
  value = rhcs_identity_provider.google_idp.id
}

output "idp_ldap_id" {
  value = rhcs_identity_provider.ldap_idp.id
}