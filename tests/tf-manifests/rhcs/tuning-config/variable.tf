variable "cluster" {
  type = string
}

variable "name_prefix" {
  type = string
}

variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "tc_count"{
  type    = number
  default = 1
}

variable "spec_vm_dirty_ratios" {
  type = list(number)
}

variable "spec_priorities" {
  type = list(number)
}
