output "machine_pools" {
  value = [ for mp in rhcs_machine_pool.mps : {
    machine_pool_id: mp.id
    name: mp.name
    cluster_id: mp.cluster
    replicas: mp.replicas
    machine_type: mp.machine_type
    autoscaling_enabled: mp.autoscaling_enabled
    labels: mp.labels
    taints: mp.taints
  } ]
}