resource "rhcs_wif_config" "wif" {
  display_name      = "my-cluster-wif"
  openshift_version = "4.21.15"
  gcp = {
    project_id     = "my-gcp-project"
    project_number = "123456789012"
    role_prefix    = "mycluster"
  }
}

resource "rhcs_cluster_osd_gcp" "osd" {
  name           = "my-cluster"
  product        = "osd"
  cloud_region   = "us-east1"
  gcp_project_id = "my-gcp-project"
  wif_config_id  = rhcs_wif_config.wif.id
  version        = "4.21.15"
  compute_nodes  = 3
  ccs_enabled    = true

  # Inline cluster-admin htpasswd user. Mirrors rhcs_cluster_rosa_classic.
  # Leave password unset to have the provider auto-generate one (exposed via
  # the admin_credentials attribute).
  create_admin_user = true
}
