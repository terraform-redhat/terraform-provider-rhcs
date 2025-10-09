resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
  cluster                       = "cluster-id-123"
  max_pod_grace_period          = 600
  pod_priority_threshold        = -10
  max_node_provision_time       = "15m"
  balance_similar_node_groups   = true
  balancing_ignored_labels      = ["example-label"]
  skip_nodes_with_local_storage = false
  log_verbosity                 = 4

  resource_limits = {
    max_nodes_total = 5
    cores = {
      max = 11520
      min = 1
    }
    memory = {
      max = 230400
      min = 1
    }
    gpu = {
      type = "nvidia.com/gpu" # or amd.com/gpu
      range = {
        min = 0
        max = 10
      }
    }
  }
}
