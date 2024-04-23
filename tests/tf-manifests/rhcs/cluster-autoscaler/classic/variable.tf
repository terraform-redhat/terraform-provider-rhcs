variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

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

variable "resource_limits" {
  type = object({
    max_nodes_total = optional(number)
    cores = object({
      min = optional(number)
      max = optional(number)
    })
    memory = object({
      min = optional(number)
      max = optional(number)
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
