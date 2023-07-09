output "oidc_config_id" {
  value = module.oidc_config.id
}

output "oidc_endpoint_url" {
  value = module.oidc_config.oidc_endpoint_url
}

output "thumbprint" {
  value = module.oidc_config.thumbprint
}

output "cluster_id" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
}
