# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "cluster_id" {
  value = rhcs_cluster_hyperfleet.cluster.id
}

output "cluster_name" {
  value = rhcs_cluster_hyperfleet.cluster.name
}

output "cluster_phase" {
  value = rhcs_cluster_hyperfleet.cluster.phase
}

output "cluster_api_url" {
  value = rhcs_cluster_hyperfleet.cluster.api_url
}

output "scheduler_profile" {
  value = rhcs_cluster_hyperfleet.cluster.scheduler_profile
}
output "autoscaler_max_node_provision_time" {
  value = rhcs_cluster_hyperfleet.cluster.autoscaler_max_node_provision_time
}
output "autoscaler_max_nodes_total" {
  value = rhcs_cluster_hyperfleet.cluster.autoscaler_max_nodes_total
}
output "autoscaler_max_pod_grace_period" {
  value = rhcs_cluster_hyperfleet.cluster.autoscaler_max_pod_grace_period
}
output "autoscaler_pod_priority_threshold" {
  value = rhcs_cluster_hyperfleet.cluster.autoscaler_pod_priority_threshold
}
output "component_routes" {
  value = rhcs_cluster_hyperfleet.cluster.component_routes
}
