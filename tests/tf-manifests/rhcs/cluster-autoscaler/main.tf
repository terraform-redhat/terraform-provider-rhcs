terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  url = var.url
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
  resource_limits = {
    max_nodes_total = var.max_nodes_total
    cores = {
      min = var.min_cores
      max = var.max_cores
    }
    memory = {
      min = var.min_memory
      max = var.max_memory
    }
  }
  scale_down = {
    enabled               = var.enabled
    utilization_threshold = var.utilization_threshold
    unneeded_time         = var.unneeded_time
    delay_after_add       = var.delay_after_add
    delay_after_delete    = var.delay_after_delete
    delay_after_failure   = var.delay_after_failure
  }
}
