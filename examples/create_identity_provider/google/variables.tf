# OCM Variables
variable "token" {
  type        = string
  description = "OCM token - You can get it here: https://console.redhat.com/openshift/token"
}

variable "cluster_id" {
  type        = string
  description = "The OCP cluster ID"
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

# IDP Variables
variable "google_client_id" {
  type        = string
  description = "Google client id"
}
variable "google_client_secret" {
  type        = string
  description = "Google client secret"
}
variable "google_hosted_domain" {
  type        = string
  description = "Restrict users to a Google Apps domain."
}
