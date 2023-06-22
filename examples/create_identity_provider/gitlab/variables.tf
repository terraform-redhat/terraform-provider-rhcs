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
variable "gitlab_client_id" {
  type        = string
  description = "GitLab client id"
}
variable "gitlab_client_secret" {
  type        = string
  description = "GitLab client secret"
}
variable "gitlab_url" {
  type        = string
  description = "Optional. The host URL of a GitLab provider. (default 'https://gitlab.com')"
  default     = "https://gitlab.com"
}
