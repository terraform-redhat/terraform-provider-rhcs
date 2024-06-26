variable "cluster" {
  type    = string
  default = null
}
variable "machine_type" {
  default = null
  type    = string
}
variable "name" {
  type    = string
  default = null
}
variable "autoscaling_enabled" {
  default = null
  type    = bool
}
variable "labels" {
  default = null
  type    = map(string)
}
variable "max_replicas" {
  type    = number
  default = null
}
variable "min_replicas" {
  type    = number
  default = null
}
variable "replicas" {
  type    = number
  default = null
}
variable "taints" {
  default = null
  type = list(object({
    key           = string
    value         = string
    schedule_type = string
  }))
}

variable "availability_zone" {
  type    = string
  default = null
}
variable "subnet_id" {
  type    = string
  default = null
}

variable "additional_security_groups" {
  type    = list(string)
  default = null
}
variable "auto_repair" {
  type    = bool
  default = null
}

variable "tags" {
  type    = map(string)
  default = null
}

variable "tuning_configs" {
  type    = list(string)
  default = []
}

variable "upgrade_acknowledgements_for" {
  type    = string
  default = null
}
variable "openshift_version" {
  type    = string
  default = null
}

variable "kubelet_configs" {
  type    = string
  default = null
}