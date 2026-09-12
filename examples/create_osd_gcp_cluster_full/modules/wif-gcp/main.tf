# Workload Identity Pool and OIDC provider for OSD WIF
# Mirrors what ocm gcp create wif-config --mode auto creates in GCP

locals {
  pool_project_id = coalesce(var.federated_project_id, var.project_id)
}

resource "google_iam_workload_identity_pool" "wif" {
  project                   = local.pool_project_id
  workload_identity_pool_id = var.pool_id
  display_name              = var.pool_id
  description               = "Created by Terraform for WIF config ${var.display_name}"
}

resource "google_iam_workload_identity_pool_provider" "oidc" {
  project                            = google_iam_workload_identity_pool.wif.project
  workload_identity_pool_id          = google_iam_workload_identity_pool.wif.workload_identity_pool_id
  workload_identity_pool_provider_id = var.identity_provider.identity_provider_id
  display_name                       = var.identity_provider.identity_provider_id
  description                        = "Created by Terraform for WIF config ${var.display_name}"
  attribute_mapping = {
    "google.subject" = "assertion.sub"
  }

  oidc {
    issuer_uri        = var.identity_provider.issuer_url
    jwks_json         = var.identity_provider.jwks
    allowed_audiences = length(var.identity_provider.allowed_audiences) > 0 ? var.identity_provider.allowed_audiences : null
  }
}
