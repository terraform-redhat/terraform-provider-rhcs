variable "cluster" {
  type    = string
  default = null
}

variable "name" {
  type        = string
  default     = null
  description = "name will be used as it is if tc_count == 1, else it will be considered as a prefix for the instance names"
}

variable "tc_count" {
  type    = number
  default = 1
}

variable "spec" {
  type        = string
  default     = null
  description = "To use if tc_count == 1"
}

variable "spec_vm_dirty_ratios" {
  type        = list(number)
  default     = null
  description = "To use if tc_count != 1"
}

variable "spec_priorities" {
  type        = list(number)
  default     = null
  description = "To use if tc_count != 1"
}

