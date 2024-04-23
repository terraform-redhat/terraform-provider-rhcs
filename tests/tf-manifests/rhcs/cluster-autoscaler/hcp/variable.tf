variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "cluster_id" {
  type = string
}

variable "max_pod_grace_period" {
  type    = number
  default = 10
}

variable "pod_priority_threshold" {
  type    = number
  default = -10
}

variable "max_node_provision_time" {
  type    = string
  default = "1h"
}

variable "resource_limits" {
  type = object({
    max_nodes_total = optional(number)
  })
  default = null
}