# AWS config
variable "aws_region" {
  type        = string
  description = "The region to create the ROSA cluster in"
}

variable "aws_shared_credentials_files" {
  type        = list(string)
  default     = null
  description = "File path to the AWS shared credentials file. This file typically used by Shared-VPC cluster."
}

# VPC config
variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
  validation {
    condition     = contains(["10.0.0.0/16", "11.0.0.0/16", "12.0.0.0/16"], var.vpc_cidr)
    error_message = "VPC CIDR limited to: 10.0.0.0/16 11.0.0.0/16 12.0.0.0/16"
  }
}

variable "availability_zones" {
  type    = list(string)
  default = null
}

variable "availability_zones_count" {
  type        = number
  default     = null
  description = "The count of availability zones to utilize within the specified AWS region, where pairs of public and private subnets are generated. Valid only when availability_zones variable is not provided."
}

variable "name_prefix" {
  type        = string
  description = "The name prefix of the vpc for resources naming"
  default     = "tf-classic"
}

variable "tags" {
  type        = map(string)
  default     = null
  description = "AWS tags to be applied to generated AWS resources of this VPC."
}
