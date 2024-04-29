terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}



resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
  cluster                 = var.cluster_id
  max_pod_grace_period    = var.max_pod_grace_period
  pod_priority_threshold  = var.pod_priority_threshold
  max_node_provision_time = var.max_node_provision_time
  resource_limits         = var.resource_limits
}
