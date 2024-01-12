output "cluster_id" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.cluster
}
output "balance_similar_node_groups" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.balance_similar_node_groups
}
output "skip_nodes_with_local_storage" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.skip_nodes_with_local_storage
}
output "log_verbosity" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.log_verbosity
}
output "max_pod_grace_period" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.max_pod_grace_period
}
output "pod_priority_threshold" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.pod_priority_threshold
}
output "ignore_daemonsets_utilization" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.ignore_daemonsets_utilization
}
output "max_node_provision_time" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.max_node_provision_time
}
output "balancing_ignored_labels" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.balancing_ignored_labels
}
output "max_nodes_total" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.resource_limits.max_nodes_total
}
output "min_cores" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.resource_limits.cores.min
}
output "max_cores" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.resource_limits.cores.max
}
output "min_memory" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.resource_limits.memory.min
}
output "max_memory" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.resource_limits.memory.max
}
output "delay_after_add" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.scale_down.delay_after_add
}
output "delay_after_delete" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.scale_down.delay_after_delete
}
output "delay_after_failure" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.scale_down.delay_after_failure
}
output "unneeded_time" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.scale_down.unneeded_time
}
output "utilization_threshold" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.scale_down.utilization_threshold
}
output "enabled" {
  value = rhcs_cluster_autoscaler.cluster_autoscaler.scale_down.enabled
}

