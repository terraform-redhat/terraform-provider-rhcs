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
