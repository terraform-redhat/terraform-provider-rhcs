variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "cluster" {
  type    = string
  default = null
}
variable "pod_pids_limit" {
  type    = number
  default = null
}

