# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "nodepool_id" {
  value = rhcs_nodepool_seaplane.nodepool.id
}

output "nodepool_name" {
  value = rhcs_nodepool_seaplane.nodepool.name
}

output "nodepool_phase" {
  value = rhcs_nodepool_seaplane.nodepool.phase
}

output "nodepool_replicas" {
  value = rhcs_nodepool_seaplane.nodepool.replicas
}

output "nodepool_min_replicas" {
  value = rhcs_nodepool_seaplane.nodepool.autoscaling == null ? null : rhcs_nodepool_seaplane.nodepool.autoscaling.min_replicas
}

output "nodepool_max_replicas" {
  value = rhcs_nodepool_seaplane.nodepool.autoscaling == null ? null : rhcs_nodepool_seaplane.nodepool.autoscaling.max_replicas
}
