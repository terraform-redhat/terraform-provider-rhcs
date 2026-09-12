# Service accounts for OSD WIF
# One per entry in the OCM blueprint

resource "google_service_account" "wif" {
  for_each = { for sa in var.service_accounts : sa.service_account_id => sa }

  project      = var.project_id
  account_id   = each.key
  display_name = "${var.display_name}-${each.key}"
  description  = "Created by Terraform for WIF config ${var.display_name}"
}
