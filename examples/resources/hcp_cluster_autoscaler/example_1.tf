resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
  cluster                 = "cluster-id-123"
  max_pod_grace_period    = 600
  pod_priority_threshold  = -10
  max_node_provision_time = "15m"

  resource_limits = {
    max_nodes_total = 5
  }
}
