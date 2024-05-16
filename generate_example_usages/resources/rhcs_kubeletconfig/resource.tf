# Example KubeletConfig
resource rhcs_kubeletconfig "example_kubeletconfig" {
  cluster = "cluster-id-123"
  pod_pids_limit = 10000
}