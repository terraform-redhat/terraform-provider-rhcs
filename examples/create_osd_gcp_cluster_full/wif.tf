###############################################################################
# WIF: OCM config + GCP-side IAM
###############################################################################
# rhcs_wif_config registers the federation in OCM and returns the GCP
# workload-identity-pool blueprint. The wif-gcp module then creates the actual
# pool, OIDC provider, service accounts, and role bindings in GCP using the
# upstream hashicorp/google provider.

resource "rhcs_wif_config" "wif" {
  display_name      = "${var.cluster_name}-wif"
  openshift_version = var.openshift_version
  gcp = {
    project_id     = var.service_project_id
    project_number = tostring(data.google_project.service.number)
    role_prefix    = replace(replace(var.cluster_name, "-", ""), "_", "")
  }
}

module "wif_gcp" {
  source = "./modules/wif-gcp"

  project_id   = var.service_project_id
  display_name = rhcs_wif_config.wif.display_name
  pool_id      = rhcs_wif_config.wif.gcp.workload_identity_pool.pool_id
  identity_provider = {
    identity_provider_id = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.identity_provider_id
    issuer_url           = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.issuer_url
    jwks                 = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.jwks
    allowed_audiences    = rhcs_wif_config.wif.gcp.workload_identity_pool.identity_provider.allowed_audiences
  }
  service_accounts         = rhcs_wif_config.wif.gcp.service_accounts
  support                  = rhcs_wif_config.wif.gcp.support
  impersonator_email       = rhcs_wif_config.wif.gcp.impersonator_email
  federated_project_id     = try(rhcs_wif_config.wif.gcp.federated_project_id, "") != "" ? rhcs_wif_config.wif.gcp.federated_project_id : null
  federated_project_number = try(rhcs_wif_config.wif.gcp.federated_project_number, "") != "" ? rhcs_wif_config.wif.gcp.federated_project_number : tostring(data.google_project.service.number)
}
