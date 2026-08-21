variable "cluster_name" {
  type        = string
  description = "Cluster name; also used to derive WIF display name and KMS key/SA names."
}

variable "openshift_version" {
  type        = string
  default     = "4.21.15"
  description = "OSD version. Must be >= 4.17 for WIF support."
}

variable "compute_nodes" {
  type        = number
  default     = 3
  description = "Worker node count. Trial quota caps the cluster at 40 worker vCPUs."
}

variable "product" {
  type        = string
  default     = "osd"
  description = "OSD product: 'osd' (paid) or 'osdtrial' (60-day evaluation)."
}

variable "billing_model" {
  type        = string
  default     = null
  description = "Billing model. Allowed: 'standard' or 'marketplace-gcp'. Leave null to let OCM pick the default for the product ('standard' for osdtrial, 'marketplace-gcp' for osd)."
}

variable "service_project_id" {
  type        = string
  description = "GCP project the cluster runs in (and where WIF/IAM is configured)."
}

variable "host_project_id" {
  type        = string
  description = "GCP project that owns the Shared VPC (host project). Pass the same as service_project_id if not using Shared VPC."
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "GCP region."
}

variable "network_name" {
  type        = string
  description = "VPC network name in the host project."
}

variable "control_plane_subnet" {
  type        = string
  description = "Subnet name for control plane nodes."
}

variable "compute_subnet" {
  type        = string
  description = "Subnet name for worker nodes."
}

variable "psc_subnet" {
  type        = string
  description = "PSC service-attachment subnet name (in the host project)."
}

variable "admin_username" {
  type        = string
  default     = "admin"
  description = "Inline cluster-admin htpasswd username."
}

variable "admin_password" {
  type        = string
  sensitive   = true
  default     = ""
  description = "Inline cluster-admin htpasswd password. Leave empty to let the provider generate one (output via the cluster's admin_credentials attribute)."
}
