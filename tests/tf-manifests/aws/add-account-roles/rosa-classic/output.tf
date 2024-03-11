output "rhcs_versions" {
  value = data.rhcs_versions.all.items
}

output "account_role_prefix" {
  value = module.create_account_roles.account_role_prefix
}

output "path" {
  value = var.path
}

output "major_version" {
  value = local.major_version
}

output "channel_group" {
  value = var.channel_group
}

output "rhcs_gateway_url" {
  value = var.url
}

output "installer_role_arn" {
  value = local.installer_role_arn
}

output "aws_account_id" {
  value = local.aws_account_id
}