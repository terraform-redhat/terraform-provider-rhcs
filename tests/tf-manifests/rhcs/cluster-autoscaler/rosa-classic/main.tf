terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0-0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
  cluster                       = var.cluster_id
  balance_similar_node_groups   = var.balance_similar_node_groups
  skip_nodes_with_local_storage = var.skip_nodes_with_local_storage
  log_verbosity                 = var.log_verbosity
  max_pod_grace_period          = var.max_pod_grace_period
  pod_priority_threshold        = var.pod_priority_threshold
  ignore_daemonsets_utilization = var.ignore_daemonsets_utilization
  max_node_provision_time       = var.max_node_provision_time
  balancing_ignored_labels      = var.balancing_ignored_labels
  resource_limits               = var.resource_limits
  scale_down                    = var.scale_down
}
