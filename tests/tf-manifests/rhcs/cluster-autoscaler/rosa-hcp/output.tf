output "cluster_id" {
  value = rhcs_hcp_cluster_autoscaler.cluster_autoscaler.cluster
}
output "max_pod_grace_period" {
  value = rhcs_hcp_cluster_autoscaler.cluster_autoscaler.max_pod_grace_period
}
output "pod_priority_threshold" {
  value = rhcs_hcp_cluster_autoscaler.cluster_autoscaler.pod_priority_threshold
}
output "max_node_provision_time" {
  value = rhcs_hcp_cluster_autoscaler.cluster_autoscaler.max_node_provision_time
}
output "max_nodes_total" {
  value = rhcs_hcp_cluster_autoscaler.cluster_autoscaler.resource_limits.max_nodes_total
}

