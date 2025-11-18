output "cluster_id" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
}

output "cluster_name" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.name
}

output "cluster_version" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.current_version
}

output "properties" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.properties
}

output "tags" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.tags
}

output "additional_compute_security_groups" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.aws_additional_compute_security_group_ids
}

output "external_auth_providers_enabled" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.external_auth_providers_enabled
}