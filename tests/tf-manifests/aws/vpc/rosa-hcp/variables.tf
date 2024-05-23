variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
  validation {
    condition     = contains(["10.0.0.0/16", "11.0.0.0/16", "12.0.0.0/16"], var.vpc_cidr)
    error_message = "VPC CIDR limited to: 10.0.0.0/16 11.0.0.0/16 12.0.0.0/16"
  }

}
variable "multi_az" {
  type    = bool
  default = false
}

variable "aws_region" {
  type        = string
  description = "The region to create the ROSA cluster in"
}

variable "az_ids" {
  type    = list(string)
  default = null
}

variable "name" {
  type        = string
  description = "The name of the vpc to create"
  default     = "tf-ocm"

}

variable "aws_shared_credentials_files" {
  type        = list(string)
  default     = null
  description = "File path to the AWS shared credentials file. This file typically used by Shared-VPC cluster."
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "AWS tags to be applied to generated AWS resources of this VPC."
}

variable "disable_subnet_tagging" {
  type = bool
  default = false
}
