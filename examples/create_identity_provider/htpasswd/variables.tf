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
variable "htpasswd_users" {
  type = list(object({
    username = string
    password = string
  }))
  description = "htpasswd user list"
}
