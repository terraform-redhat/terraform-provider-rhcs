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

output "nodepool_placement_market_type" {
  value = rhcs_nodepool_seaplane.nodepool.placement_market_type == null ? "" : rhcs_nodepool_seaplane.nodepool.placement_market_type
}

output "nodepool_max_price" {
  value = rhcs_nodepool_seaplane.nodepool.max_price == null ? "" : rhcs_nodepool_seaplane.nodepool.max_price
}

output "nodepool_min_replicas" {
  value = rhcs_nodepool_seaplane.nodepool.autoscaling == null ? null : rhcs_nodepool_seaplane.nodepool.autoscaling.min_replicas
}

output "nodepool_max_replicas" {
  value = rhcs_nodepool_seaplane.nodepool.autoscaling == null ? null : rhcs_nodepool_seaplane.nodepool.autoscaling.max_replicas
}
