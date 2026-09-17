output "id" {
  value = module.oidc_config_and_provider.oidc_config_id
}

output "oidc_endpoint_url" {
  value = module.oidc_config_and_provider.oidc_endpoint_url
}
