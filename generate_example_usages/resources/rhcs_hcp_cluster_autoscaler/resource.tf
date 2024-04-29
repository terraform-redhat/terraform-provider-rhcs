resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
  cluster                 = "cluster-id-123"
  max_pod_grace_period    = 15
  pod_priority_threshold  = 1
  max_node_provision_time = 10

  resource_limits = {
    max_nodes_total = 5
  }
}
