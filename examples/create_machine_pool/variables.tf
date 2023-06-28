variable "token" {
  type = string
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

variable "cluster_id" {
  description = "The ID of the cluster which the machine pool is created for."
  type        = string
}

variable "name" {
  description = "The machine pool name."
  type        = string
}

variable "machine_type" {
  description = "The AWS instance type that used for the instances creation ."
  type        = string
}

variable "replicas" {
  description = "The amount of the machine created in this machine pool."
  type        = number
}
