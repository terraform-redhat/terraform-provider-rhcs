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
variable "github_client_id" {
  type        = string
  description = "GitHub client id"
}
variable "github_client_secret" {
  type        = string
  description = "GitHub client secret"
}
variable "github_org" {
  type        = string
  description = "GitHub organization"
}
