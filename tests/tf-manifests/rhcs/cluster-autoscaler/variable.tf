variable "cluster_id" {
  type = string
}
variable "balance_similar_node_groups" {
  default = true
  type    = bool
}
variable "skip_nodes_with_local_storage" {
  default = true
  type    = bool
}
variable "log_verbosity" {
  type    = number
  default = 1
}
variable "max_pod_grace_period" {
  type    = number
  default = 10
}
variable "pod_priority_threshold" {
  type    = number
  default = -10
}
variable "ignore_daemonsets_utilization" {
  default = true
  type    = bool
}
variable "max_node_provision_time" {
  type    = string
  default = "1h"
}
variable "balancing_ignored_labels" {
  default = null
  type    = list(string)
}
variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "max_nodes_total" {
  type    = number
  default = 10
}
variable "min_cores" {
  type    = number
  default = 0
}
variable "max_cores" {
  type    = number
  default = 1
}
variable "min_memory" {
  type    = number
  default = 0
}
variable "max_memory" {
  type    = number
  default = 1
}
variable "enabled" {
  default = true
  type    = bool
}
variable "utilization_threshold" {
  type    = string
  default = "0.5"
}
variable "unneeded_time" {
  type    = string
  default = "1h"
}
variable "delay_after_add" {
  type    = string
  default = "1h"
}
variable "delay_after_delete" {
  type    = string
  default = "1h"
}
variable "delay_after_failure" {
  type    = string
  default = "1h"
}
variable "resource_limits" {
  type = object({
    max_nodes_total = optional(number)
    cores = object({
      min_cores = optional(number)
      max_cores = optional(number)
    })
    memory = object({
      min_memory = optional(number)
      max_memory = optional(number)
    })
  })
  default = null
}
variable "scale_down" {
  type = object({
    enabled               = bool
    utilization_threshold = optional(string)
    unneeded_time         = optional(string)
    delay_after_add       = optional(string)
    delay_after_delete    = optional(string)
    delay_after_failure   = optional(string)
  })
  default = null
}
