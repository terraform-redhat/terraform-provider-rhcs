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

variable "installer_role_arn" {
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
  type    = list(string)
  default = ["us-east-2a"]
}
