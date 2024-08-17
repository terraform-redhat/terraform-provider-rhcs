variable "shared_vpc_aws_shared_credentials_files" {
  type        = list(string)
  description = "List of files path to the AWS shared credentials file. This file typically contains AWS access keys and secret keys and is used when authenticating with AWS using profiles (default file located at ~/.aws/credentials)."
}

variable "region" {
  type        = string
  description = "List of files path to the AWS shared credentials file. This file typically contains AWS access keys and secret keys and is used when authenticating with AWS using profiles (default file located at ~/.aws/credentials)."
}

variable "domain_prefix" {
  type        = string
  description = "The domain prefix used for hosted zone creation. It's utilized for the Hosted Zone domain."
  default     = null
}

variable "cluster_name" {
  type        = string
  description = "The cluster's name for which shared resources are created. It's utilized for the Hosted Zone domain when domain_prefix not set."
}
variable "dns_domain_id" {
  type        = string
  description = "The Base Domain that should be used for the Hosted Zone creation."
}

variable "ingress_operator_role_arn" {
  type        = string
  description = "Ingress Operator ARN from target account"
}

variable "installer_role_arn" {
  type        = string
  description = "Installer ARN from target account"
}

variable "cluster_aws_account" {
  type        = string
  description = "The AWS account number in where the cluster is going to be created."
}

variable "vpc_id" {
  type        = string
  description = "The Shared VPC ID"
}

variable "subnets" {
  type        = list(string)
  description = "The list of the subnets that should be shared between the accounts."
}
