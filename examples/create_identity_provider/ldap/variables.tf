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
variable "ldap_ca" {
  type        = string
  description = "Optional trusted certificate authority bundle"
}
variable "ldap_insecure" {
  type        = bool
  description = "Do not make TLS connections to the server."
  default     = false
}
variable "ldap_url" {
  type        = string
  description = "An RFC 2255 URL which specifies the LDAP search parameters to use."
}
