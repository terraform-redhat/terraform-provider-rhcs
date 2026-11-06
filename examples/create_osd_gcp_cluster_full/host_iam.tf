###############################################################################
# Host-project IAM grants for OSD Shared-VPC cluster SAs
###############################################################################
# OSD validates that osd-deployer, osd-control-plane, and machine-api-gcp
# (created in the service project by the wif-gcp module) have
# Compute Network Admin + Compute Security Admin + DNS Administrator
# on the Shared VPC host project. Without these the cluster sits in
# `waiting` with "Could not validate the shared subnets".

locals {
  host_required_sa_prefixes = [
    "osd-deployer-",
    "osd-control-plane-",
    "machine-api-gcp-",
  ]

  host_iam_principals = [
    for sa in rhcs_wif_config.wif.gcp.service_accounts :
    "serviceAccount:${sa.service_account_id}@${var.service_project_id}.iam.gserviceaccount.com"
    if anytrue([for p in local.host_required_sa_prefixes : startswith(sa.service_account_id, p)])
  ]

  host_iam_roles = [
    "roles/compute.networkAdmin",
    "roles/compute.securityAdmin",
    "roles/dns.admin",
  ]

  host_iam_bindings = {
    for pair in setproduct(local.host_iam_principals, local.host_iam_roles) :
    "${pair[0]}-${pair[1]}" => { member = pair[0], role = pair[1] }
  }
}

resource "google_project_iam_member" "host" {
  for_each = local.host_iam_bindings

  project = var.host_project_id
  role    = each.value.role
  member  = each.value.member

  depends_on = [module.wif_gcp]
}
