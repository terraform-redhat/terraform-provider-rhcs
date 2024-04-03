variable "cluster" {
  type = string
}
variable "machine_type" {
  default = null
  type    = string
}
variable "name" {
  type = string
}
variable "autoscaling_enabled" {
  default = false
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
variable "max_spot_price" {
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
variable "use_spot_instances" {
  type    = bool
  default = false
}
variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "availability_zone" {
  type    = string
  default = null
}
variable "subnet_id" {
  type    = string
  default = null
}
variable "multi_availability_zone" {
  type    = bool
  default = null
}
variable "disk_size" {
  type    = number
  default = null
}
variable "additional_security_groups" {
  type    = list(string)
  default = null
}
