output "idp_google_id" {
  value = rhcs_identity_provider.google_idp.id
}

output "idp_ldap_id" {
  value = rhcs_identity_provider.ldap_idp.id
}