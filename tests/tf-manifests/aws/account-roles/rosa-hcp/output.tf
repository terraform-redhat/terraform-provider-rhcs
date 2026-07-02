# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "account_role_prefix" {
  value = module.create_account_roles.account_role_prefix
}

output "installer_role_arn" {
  value = module.create_account_roles.account_roles_arn["HCP-ROSA-Installer"]
}

output "aws_account_id" {
  value = local.aws_account_id
}

output "account_roles_arn" {
  value = module.create_account_roles.account_roles_arn
}

output "trust_policy_external_id" {
  value = module.create_account_roles.trust_policy_external_id
}
