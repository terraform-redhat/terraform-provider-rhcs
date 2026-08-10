# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "cluster_id" {
  value = rhcs_cluster_hyperfleet.cluster.id
}

output "cluster_name" {
  value = rhcs_cluster_hyperfleet.cluster.name
}

output "cluster_phase" {
  value = rhcs_cluster_hyperfleet.cluster.phase
}

output "cluster_api_url" {
  value = rhcs_cluster_hyperfleet.cluster.api_url
}

output "oidc_issuer" {
  value = rhcs_cluster_hyperfleet.cluster.oidc_issuer
}
