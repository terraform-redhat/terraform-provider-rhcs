# Example KubeletConfig
resource rhcs_kubeletconfig "example_kubeletconfig" {
  cluster = "<cluster-id>"
  pod_pids_limit = 10000
}