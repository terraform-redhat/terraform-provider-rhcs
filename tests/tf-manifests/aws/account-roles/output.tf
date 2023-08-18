output "rhcs_versions" {
  value = data.rhcs_versions.all.items[0].name
}