output "rhcs_versions" {
  value = data.rhcs_versions.all.items
}

output "account_roles_prefix" {
  value = module.create_account_roles.account_role_prefix
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
