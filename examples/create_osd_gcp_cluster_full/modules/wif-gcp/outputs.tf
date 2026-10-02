output "workload_identity_pool_name" {
  description = "Full resource name of the workload identity pool"
  value       = google_iam_workload_identity_pool.wif.name
}

output "workload_identity_pool_id" {
  description = "Workload identity pool ID"
  value       = google_iam_workload_identity_pool.wif.workload_identity_pool_id
}

output "workload_identity_pool_provider_name" {
  description = "Full resource name of the OIDC provider"
  value       = google_iam_workload_identity_pool_provider.oidc.name
}

output "service_account_emails" {
  description = "Map of service account ID to email for created service accounts"
  value       = { for k, v in google_service_account.wif : k => v.email }
}
