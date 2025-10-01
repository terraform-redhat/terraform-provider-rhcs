# Cluster Configuration
variable "cluster_name" {
  type        = string
  description = "Name of the ROSA HCP cluster"
}

variable "aws_region" {
  type        = string
  description = "AWS region where the cluster will be created"
  default     = "us-east-1"
}

variable "aws_account_id" {
  type        = string
  description = "AWS account ID"
}

variable "aws_billing_account_id" {
  type        = string
  description = "AWS billing account ID (usually same as aws_account_id)"
}

variable "subnet_ids" {
  type        = list(string)
  description = "List of subnet IDs for the cluster"
}

variable "availability_zones" {
  type        = list(string)
  description = "List of availability zones for the cluster"
}

variable "replicas" {
  type        = number
  description = "Number of worker nodes"
  default     = 2
}

variable "openshift_version" {
  type        = string
  description = "OpenShift version for the cluster"
  default     = "4.15.9"
}

# STS Role Configuration
variable "installer_role_arn" {
  type        = string
  description = "ARN of the installer role"
}

variable "support_role_arn" {
  type        = string
  description = "ARN of the support role"
}

variable "worker_role_arn" {
  type        = string
  description = "ARN of the worker role"
}

variable "operator_role_prefix" {
  type        = string
  description = "Prefix for operator roles"
}

variable "oidc_config_id" {
  type        = string
  description = "OIDC configuration ID"
}

# Optional Cluster Configuration
variable "cluster_properties" {
  type        = map(string)
  description = "Additional cluster properties"
  default     = {}
}

variable "cluster_tags" {
  type        = map(string)
  description = "Tags to apply to the cluster"
  default     = {}
}

# Image Mirrors Configuration
variable "image_mirrors" {
  type        = map(list(string))
  description = "Map of source registries to their mirror lists"
  default = {
    "docker.io/library/nginx" = [
      "quay.io/my-org/nginx",
      "registry.example.com/nginx"
    ]
    "docker.io/library/redis" = [
      "quay.io/my-org/redis",
      "registry.example.com/redis"
    ]
    "docker.io/library/postgres" = [
      "registry.example.com/postgres",
      "quay.io/backup/postgres"
    ]
    "quay.io/prometheus/prometheus" = [
      "registry.corp.example.com/prometheus/prometheus"
    ]
  }
}