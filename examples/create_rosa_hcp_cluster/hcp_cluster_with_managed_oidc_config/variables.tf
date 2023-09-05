variable "token" {
  type = string
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

variable "cloud_region" {
  description = <<-EOT
    Region to create AWS infrastructure resources for a
      ROSA with hosted control planes cluster. (required)
  EOT
  type    = string
  default = "us-east-1"
}

variable "availability_zones" {
  type    = list(string)
  default = ["us-east-1a"]
}

variable "aws_subnet_ids" {
  description = "List of AWS VPC Subnet IDs for the cluster."
  type    = list(string)
}

variable "cluster_name" {
  description = "Name of the created ROSA with hosted control planes cluster."
  type        = string
  default     = "rosa-hcp"

  validation {
    condition     = can(regex("^[a-z][-a-z0-9]{0,13}[a-z0-9]$", var.cluster_name))
    error_message = <<-EOT
      ROSA cluster names must be less than 16 characters.
        May only contain lower case, alphanumeric, or hyphens characters.
    EOT
  }
}

variable "managed" {
  description = "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted) OIDC Configuration"
  type        = bool
  default     = true
}

variable "account_role_prefix" {
  type    = string
}

variable "operator_role_prefix" {
  type = string
}

variable "installer_role_arn" {
  description = "STS Role ARN with get secrets permission, relevant only for unmanaged OIDC config"
  type        = string
  default     = null
}

variable "rh_oidc_provider_thumbprint" {
  description = "Thumbprint for https://rh-oidc.s3.us-east-1.amazonaws.com"
  type        = string
  default     = "917e732d330f9a12404f73d8bea36948b929dffc"
}

variable "replicas" {
  description = "The amount of the machine created in this machine pool."
  type        = number
  default     = 2
}

variable "openshift_version" {
  description = "Desired version of OpenShift for the cluster, for example '4.1.0'. If version is greater than the currently running version, an upgrade will be scheduled."
  type        = string
  default     = "4.13.8"
}

variable "username" {
  type        = string
  description = "The cluster admin username"
  default     = "cluster-admin"
}

variable "ocm_environment" {
  type    = string
  default = "production"
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}
