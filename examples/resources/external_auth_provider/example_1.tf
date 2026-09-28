variable "cluster_id" {
  description = "ID of a ready ROSA HCP cluster with external authentication enabled."
  type        = string
}

variable "console_client_id" {
  description = "OIDC client ID registered for the OpenShift console."
  type        = string
}

variable "console_client_secret" {
  description = "OIDC client secret for the OpenShift console."
  type        = string
  sensitive   = true
}

variable "issuer_url" {
  description = "HTTPS URL of the OIDC token issuer."
  type        = string
}

resource "rhcs_external_auth_provider" "main" {
  cluster          = var.cluster_id
  name             = "example"
  issuer_url       = var.issuer_url
  issuer_audiences = [var.console_client_id]
  # Optional: provide PEM certificate content when the issuer uses a private CA.
  # issuer_ca = file("ca.pem")
  console_client_id     = var.console_client_id
  console_client_secret = var.console_client_secret
  # Request only if this identity provider requires the profile scope.
  console_extra_scopes = ["profile"]
}
