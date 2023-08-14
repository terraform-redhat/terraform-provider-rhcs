variable "token" {
  type = string
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

variable "cluster_id" {
  description = "The ID of the cluster which the cluster-autoscaler is created for."
  type        = string
}
