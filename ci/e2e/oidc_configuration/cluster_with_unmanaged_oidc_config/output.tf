output "oidc_config_id" {
  value = rhcs_rosa_oidc_config.oidc_config.id
}

output "oidc_endpoint_url" {
  value = rhcs_rosa_oidc_config.oidc_config.oidc_endpoint_url
}

output "thumbprint" {
  value = rhcs_rosa_oidc_config.oidc_config.thumbprint
}

output "cluster_id" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
}

