output "versions" {
  description = "OpenShift versions"
  value       = data.rhcs_versions.all
}
