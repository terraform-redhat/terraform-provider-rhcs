output "cluster_id" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
}

output "oidc_thumbprint" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.sts.thumbprint
}

output "oidc_endpoint_url" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url
}

output "cluster_admin_username" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.admin_credentials == null ? null : rhcs_cluster_rosa_classic.rosa_sts_cluster.admin_credentials.username
}

output "cluster_admin_password" {
  value     = rhcs_cluster_rosa_classic.rosa_sts_cluster.admin_credentials == null ? null : rhcs_cluster_rosa_classic.rosa_sts_cluster.admin_credentials.password
  sensitive = true
}