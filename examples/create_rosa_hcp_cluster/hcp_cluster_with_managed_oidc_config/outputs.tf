output "cloud_region" {
  value = var.cloud_region
}

output "account_role_prefix" {
  value = module.create_account_roles.account_role_prefix
}

output "cluster_id" {
  value = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
}

output "cluster_name" {
  value = var.cluster_name
}

output "api_url" {
  value = data.rhcs_cluster_data.cluster.api_url
}

output "console_url" {
  value = data.rhcs_cluster_data.cluster.console_url
}

output "cluster_admin_username" {
  value = var.username
}

output "openshift_version" {
  value = var.openshift_version
}

