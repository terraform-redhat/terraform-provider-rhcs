data "rhcs_wif_config" "wif" {
  display_name = "my-cluster-wif"
}

output "wif_pool_id" {
  value = data.rhcs_wif_config.wif.gcp.workload_identity_pool.pool_id
}
