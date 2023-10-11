# example for unmanaged oidc
resource "rhcs_rosa_oidc_config" "oidc_config" {
  managed            = false
  secret_arn         = "<secret-arn>"
  issuer_url         = "<insecure-url>"
  installer_role_arn = "<installer-role-arn>"
}

# example for managed oidc
resource "rhcs_rosa_oidc_config" "oidc_config" {
  managed            = true
}
