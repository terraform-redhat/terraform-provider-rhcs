output "account_role_prefix" {
  value = module.create_account_roles.account_role_prefix
}

output "installer_role_arn" {
  value = module.create_account_roles.account_roles_arn["Installer"]
}

output "aws_account_id" {
  value = local.aws_account_id
}
