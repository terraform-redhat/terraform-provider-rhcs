output "operator_role_prefix" {
  value = module.operator_roles.operator_role_prefix
}

output "account_role_prefix" {
  value = var.account_role_prefix
}

output "oidc_config_id" {
  value = module.oidc_config_and_provider.oidc_config_id
}

output "ingress_operator_role_arn" {
  value = local.ingress_operator_role_arn
}
