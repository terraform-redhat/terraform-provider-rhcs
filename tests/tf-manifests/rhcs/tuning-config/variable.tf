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

variable "specs" {
  type = list(object({
    spec_type  = string
    spec_value = string
  }))
  description = "List of spec objects (spec_type = [\"file\" or \"string\"] and spec_value) to use. Length of that list should be the same as tc_count"
}
