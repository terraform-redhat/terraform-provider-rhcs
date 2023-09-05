output "secret_arn" {
  value = var.create_oidc_config_resources ? module.rosa_oidc_config_resources[0].secret_arn : null
}

output "account_role_prefix" {
  value = var.create_account_roles && length(module.rosa_account_roles) > 0 ? module.rosa_account_roles[0].account_role_prefix : null
}
