variable "aws_region" {
  type = string
}
variable "sg_number" {
  type    = number
  default = null
}

variable "vpc_id" {
  type    = string
  default = null
}

variable "name_prefix" {
  type    = string
  default = "rhcs-ci"
}

variable "description" {
  type    = string
  default = "Testing for rhcs provider CI"
}