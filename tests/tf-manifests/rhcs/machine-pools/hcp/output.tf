output "machine_pools" {
  value = [for mp in rhcs_hcp_machine_pool.mps : {
    machine_pool_id : mp.id
    name : mp.name
    cluster_id : mp.cluster
    replicas : mp.replicas
    machine_type : mp.aws_node_pool.instance_type
    autoscaling_enabled : mp.autoscaling.enabled
    labels : mp.labels
    taints : mp.taints
    tuning_configs : mp.tuning_configs
    kubelet_configs : mp.kubelet_configs
    ec2_metadata_http_tokens : mp.aws_node_pool.ec2_metadata_http_tokens
  }]
}