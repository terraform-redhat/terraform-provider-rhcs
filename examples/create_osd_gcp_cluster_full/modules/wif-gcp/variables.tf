variable "project_id" {
  type        = string
  description = "GCP project ID where WIF service accounts and IAM bindings are created"
}

variable "display_name" {
  type        = string
  description = "WIF config display name (used in resource descriptions)"
}

variable "pool_id" {
  type        = string
  description = "Workload identity pool ID from OCM (rhcs_wif_config.gcp.workload_identity_pool.pool_id)"
}

variable "identity_provider" {
  type = object({
    identity_provider_id = string
    issuer_url           = string
    jwks                 = string
    allowed_audiences    = list(string)
  })
  description = "OIDC identity provider configuration from OCM"
}

variable "service_accounts" {
  type = list(object({
    service_account_id = string
    access_method      = string
    osd_role           = string
    roles = list(object({
      role_id     = string
      predefined  = bool
      permissions = list(string)
      resource_bindings = optional(list(object({
        type = string
        name = string
      })), [])
    }))
    credential_request = optional(object({
      namespace             = string
      service_account_names = list(string)
    }), null)
  }))
  description = "Service accounts blueprint from OCM (rhcs_wif_config.gcp.service_accounts)"
  default     = []
}

variable "support" {
  type = object({
    principal = string
    roles = list(object({
      role_id     = string
      predefined  = bool
      permissions = list(string)
      resource_bindings = optional(list(object({
        type = string
        name = string
      })), [])
    }))
  })
  description = "Support access configuration from OCM (rhcs_wif_config.gcp.support)"
  default     = null
}

variable "impersonator_email" {
  type        = string
  description = "Service account email used by OCM for impersonation (rhcs_wif_config.gcp.impersonator_email)"
  default     = ""
}

variable "federated_project_id" {
  type        = string
  description = "GCP project ID for workload identity pool (if different from project_id)"
  default     = null
}

variable "federated_project_number" {
  type        = string
  description = "GCP project number for the project hosting the workload identity pool (used for WIF principal construction)"
}
