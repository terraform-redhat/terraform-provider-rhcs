output "oidc_config_id" {
  value       = module.oidc_config_and_provider.oidc_config_id
  description = "The unique identifier associated with users authenticated through OpenID Connect (OIDC) generated by this OIDC config."
}

output "operator_role_prefix" {
  value       = module.operator_roles.operator_role_prefix
  description = "Prefix used for generated AWS operator policies."
}

output "account_role_prefix" {
  value = var.account_role_prefix
}

output "ingress_operator_role_arn" {
  value = module.operator_roles.operator_roles_arn["openshift-ingress-operator"]
}
