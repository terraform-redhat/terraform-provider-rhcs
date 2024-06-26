output "cluster_id" {
  value = rhcs_kubeletconfig.kubeletconfig.cluster
}

output "kubelet_configs" {
  value = rhcs_kubeletconfig.kubeletconfig
}
