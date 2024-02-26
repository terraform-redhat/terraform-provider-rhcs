output "rhcs_versions" {
  value = data.rhcs_versions.all_versions.items
}

output "account_role_prefix" {
  value = local.account_role_prefix_valid
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
