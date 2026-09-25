resource "rhcs_wif_config" "wif" {
  display_name      = "my-cluster-wif"
  openshift_version = "4.21.15"
  gcp = {
    project_id     = "my-gcp-project"
    project_number = "123456789012"
    role_prefix    = "mycluster"
  }
}
