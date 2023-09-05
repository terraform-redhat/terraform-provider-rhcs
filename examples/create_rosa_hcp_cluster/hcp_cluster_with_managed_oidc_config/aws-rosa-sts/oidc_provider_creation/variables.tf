variable "cluster_id" {
  description = "cluster ID"
  type        = string
  default     = ""
}

variable "rh_oidc_provider_url" {
  description = "oidc provider url"
  type        = string
  default     = "rh-oidc.s3.us-east-1.amazonaws.com"
}

variable "rh_oidc_provider_thumbprint" {
  description = "Thumbprint for the variable `rh_oidc_provider_url`"
  type        = string
  default     = "917e732d330f9a12404f73d8bea36948b929dffc"
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}
