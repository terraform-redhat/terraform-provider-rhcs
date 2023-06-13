variable "token" {
  type      = string
  sensitive = true
}

variable "url" {
  type    = string
  default = "https://api.openshift.com"
}

variable "operator_role_prefix" {
  type = string
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "cluster_name" {
  type    = string
  default = "tf-gdb-test"
}

variable "cloud_region" {
  type    = string
  default = "us-east-2"
}

variable "availability_zones" {
  type    = list(string***REMOVED***
  default = ["us-east-2a"]
}

variable "account_role_path" {
  description = "(Optional***REMOVED*** Path to the account role."
  type        = string
  default     = "/"
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string***REMOVED***
  default     = null
}

