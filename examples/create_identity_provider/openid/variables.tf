# OCM Variables
variable "token" {
  type        = string
  description = "OCM token - You can get it here: https://console.redhat.com/openshift/token."
}

variable "cluster_id" {
  type        = string
  description = "The OCP cluster ID."
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url."
  default     = "https://api.openshift.com"
}

# IDP Variables
variable "openid_client_id" {
  type        = string
  description = "Client ID from the registered application."
}
variable "openid_client_secret" {
  type        = string
  description = "Client Secret from the registered application."
  sensitive   = true
}
variable "openid_issuer" {
  type        = string
  description = "The URL that the OpenID Provider asserts as the Issuer Identifier. It must use the https scheme with no URL query parameters or fragment."
}
variable "openid_ca" {
  type        = string
  description = "Optional trusted certificate authority bundle, in one liner format use /n."
}
variable "openid_claims" {
  description = "Claim fields, at least one must be configured."
  type        = map(list(string))
}
