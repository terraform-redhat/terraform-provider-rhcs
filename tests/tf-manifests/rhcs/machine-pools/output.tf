output "machine_pool_id" {
  value = rhcs_machine_pool.mp.id
}
output "name" {
  value = rhcs_machine_pool.mp.name
}
output "cluster_id" {
  value = rhcs_machine_pool.mp.cluster
}

output "replicas" {
  value = rhcs_machine_pool.mp.replicas
}

output "machine_type" {
  value = rhcs_machine_pool.mp.machine_type
}

output "autoscaling_enabled" {
  value = rhcs_machine_pool.mp.autoscaling_enabled
}

output "labels" {
  value = rhcs_machine_pool.mp.labels
}

output "taints" {
  value = rhcs_machine_pool.mp.taints
}