resource "rhcs_gcp_machine_pool" "large" {
  cluster_id    = rhcs_cluster_osd_gcp.osd.id
  name          = "large"
  instance_type = "n2-standard-8"
  replicas      = 2

  labels = {
    "node-role.kubernetes.io/large" = "true"
  }

  gcp = {
    secure_boot = true
  }
}
