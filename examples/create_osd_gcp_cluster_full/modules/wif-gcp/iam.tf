# IAM resources for OSD WIF: custom roles, project bindings, SA-level bindings, support access

locals {
  # Deduplicated custom roles (role_id => role)
  custom_roles_map = {
    for role in flatten([
      for sa in var.service_accounts : [
        for r in sa.roles : r if !r.predefined
      ]
    ]) : role.role_id => role
  }

  # Project-level role bindings: (sa_id, role) where role has no resource_bindings
  project_bindings = flatten([
    for sa in var.service_accounts : [
      for r in sa.roles : {
        sa_id      = sa.service_account_id
        role_id    = r.role_id
        predefined = r.predefined
      } if r.resource_bindings == null || length(coalesce(r.resource_bindings, [])) == 0
    ]
  ])

  # WIF principals: (target_sa, namespace, sa_name) for workloadIdentityUser
  wif_bindings = flatten([
    for sa in var.service_accounts : [
      for sa_name in coalesce(try(sa.credential_request.service_account_names, []), []) : {
        target_sa = sa.service_account_id
        namespace = try(sa.credential_request.namespace, "")
        sa_name   = sa_name
      }
    ] if sa.access_method == "wif" && sa.credential_request != null
  ])

  # Resource-specific bindings: (target_sa, member_sa, role_id, predefined)
  resource_bindings = flatten([
    for sa in var.service_accounts : [
      for r in sa.roles : [
        for b in coalesce(r.resource_bindings, []) : {
          target_sa    = b.name
          member_sa    = sa.service_account_id
          role_id      = r.role_id
          predefined   = r.predefined
          binding_type = b.type
        } if b.type == "iam.serviceAccounts"
      ]
    ]
  ])

  # Support role bindings (when support is configured)
  support_bindings = var.support != null ? [
    for r in var.support.roles : {
      role_id    = r.role_id
      predefined = r.predefined
    }
  ] : []

  sa_email = { for k, v in google_service_account.wif : k => v.email }
}

# Custom IAM roles
resource "google_project_iam_custom_role" "wif" {
  for_each = local.custom_roles_map

  project     = var.project_id
  role_id     = each.key
  title       = each.key
  description = "Created by Terraform for Workload Identity Federation on OpenShift"
  permissions = each.value.permissions
}

# Project-level IAM bindings for service accounts
resource "google_project_iam_member" "sa_roles" {
  for_each = {
    for b in local.project_bindings : "${b.sa_id}-${b.role_id}" => b
  }

  project = var.project_id
  role    = each.value.predefined ? "roles/${each.value.role_id}" : google_project_iam_custom_role.wif[each.value.role_id].id
  member  = "serviceAccount:${local.sa_email[each.value.sa_id]}"
}

# Impersonator access (serviceAccountTokenCreator) for SAs with access_method=impersonate
resource "google_service_account_iam_member" "impersonator" {
  for_each = {
    for sa in var.service_accounts : sa.service_account_id => sa
    if sa.access_method == "impersonate" && var.impersonator_email != ""
  }

  service_account_id = google_service_account.wif[each.key].name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${var.impersonator_email}"
}

# WIF access (workloadIdentityUser) for SAs with access_method=wif
resource "google_service_account_iam_member" "wif_principal" {
  for_each = {
    for i, b in local.wif_bindings : "${b.target_sa}-${b.namespace}-${b.sa_name}" => b
    if b.namespace != "" && b.sa_name != ""
  }

  service_account_id = google_service_account.wif[each.value.target_sa].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/projects/${var.federated_project_number}/locations/global/workloadIdentityPools/${google_iam_workload_identity_pool.wif.workload_identity_pool_id}/subject/system:serviceaccount:${each.value.namespace}:${each.value.sa_name}"
}

# Resource-specific bindings (e.g. iam.serviceAccountUser on another SA)
# Target SA may be one we create or an external SA (by account_id)
resource "google_service_account_iam_member" "resource_binding" {
  for_each = {
    for i, b in local.resource_bindings : "${b.target_sa}-${b.member_sa}-${b.role_id}" => b
  }

  service_account_id = contains(keys(google_service_account.wif), each.value.target_sa) ? google_service_account.wif[each.value.target_sa].name : "projects/${var.project_id}/serviceAccounts/${each.value.target_sa}@${var.project_id}.iam.gserviceaccount.com"
  role               = each.value.predefined ? "roles/${each.value.role_id}" : google_project_iam_custom_role.wif[each.value.role_id].id
  member             = "serviceAccount:${local.sa_email[each.value.member_sa]}"
}

# Support access: custom roles
resource "google_project_iam_custom_role" "support" {
  for_each = var.support != null ? {
    for r in var.support.roles : r.role_id => r
    if !r.predefined
  } : {}

  project     = var.project_id
  role_id     = each.key
  title       = each.key
  description = "Created by Terraform for Workload Identity Federation on OpenShift"
  permissions = each.value.permissions
}

# Support access: project-level bindings
resource "google_project_iam_member" "support" {
  for_each = var.support != null ? {
    for r in var.support.roles : r.role_id => r
  } : {}

  project = var.project_id
  role    = each.value.predefined ? "roles/${each.value.role_id}" : google_project_iam_custom_role.support[each.value.role_id].id
  member  = "group:${var.support.principal}"
}
