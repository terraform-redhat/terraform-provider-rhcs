variable "aws_region" {
  type        = string
  description = "aws region in use"
}

variable "vpc_id" {
  type        = string
  description = " vpc id"
  default     = null
}


variable "subnet_public_id" {
  type        = string
  description = "public subnet id"
  default     = null
}


variable "trust_bundle_path" {
  type        = string
  description = "the file path of the trust bundle"
  default     = null
}

variable "key_pair_id" {
  type = string
  description = "used for key pair name. Default will be randomly generated"
  default     = null
}
