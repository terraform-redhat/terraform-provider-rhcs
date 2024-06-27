output "cluster_id" {
  value = rhcs_kubeletconfig.kubeletconfig.cluster
}

output "pod_pids_limit" {
  value = rhcs_kubeletconfig.kubeletconfig.pod_pids_limit
}