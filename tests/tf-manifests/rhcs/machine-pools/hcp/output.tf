output "machine_pool_id" {
  value = rhcs_hcp_machine_pool.mp.id
}
output "name" {
  value = rhcs_hcp_machine_pool.mp.name
}
output "cluster_id" {
  value = rhcs_hcp_machine_pool.mp.cluster
}

output "replicas" {
  value = rhcs_hcp_machine_pool.mp.replicas
}

output "machine_type" {
  value = rhcs_hcp_machine_pool.mp.aws_node_pool.instance_type
}

output "autoscaling_enabled" {
  value = rhcs_hcp_machine_pool.mp.autoscaling
}

output "labels" {
  value = rhcs_hcp_machine_pool.mp.labels
}

output "taints" {
  value = rhcs_hcp_machine_pool.mp.taints
}

output "tuning_configs" {
  value = rhcs_hcp_machine_pool.mp.tuning_configs
}

output "kubelet_configs" {
  value = rhcs_hcp_machine_pool.mp.kubelet_configs
}