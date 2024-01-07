data "rhcs_machine_pool" "machine_pool" {
  cluster = var.cluster_id
  id = var.machine_pool_id
}
