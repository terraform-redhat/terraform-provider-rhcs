# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "nodepool_id" {
  value = rhcs_nodepool_hyperfleet.nodepool.id
}

output "nodepool_name" {
  value = rhcs_nodepool_hyperfleet.nodepool.name
}

output "nodepool_phase" {
  value = rhcs_nodepool_hyperfleet.nodepool.phase
}

output "nodepool_replicas" {
  value = rhcs_nodepool_hyperfleet.nodepool.replicas
}

output "nodepool_min_replicas" {
  value = rhcs_nodepool_hyperfleet.nodepool.autoscaling == null ? null : rhcs_nodepool_hyperfleet.nodepool.autoscaling.min_replicas
}

output "nodepool_max_replicas" {
  value = rhcs_nodepool_hyperfleet.nodepool.autoscaling == null ? null : rhcs_nodepool_hyperfleet.nodepool.autoscaling.max_replicas
}
