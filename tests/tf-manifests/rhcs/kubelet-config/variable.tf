variable "cluster" {
  type    = string
  default = null
}
variable "pod_pids_limit" {
  type    = number
  default = null
}
variable "name_prefix" {
  type        = string
  default     = null
  description = "the name prefix of the kubeletconfig. When it is set, the kubeletconfig name will be like {name_prefix}-{count.index}"
}
variable "kubelet_config_number" {
  type    = number
  default = 1
}

