# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API base URL"
}

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "cluster_name" {
  type        = string
  description = "Human-readable cluster name"
}

variable "release_image" {
  type        = string
  description = "HostedCluster release image pullspec (not an OpenShift version label)"
}

variable "operator_roles_prefix" {
  type        = string
  description = "Prefix for the seven operator IAM roles"
}

variable "subnet_id" {
  type        = string
  description = "Worker-node subnet ID"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID containing the worker subnet"
}

variable "availability_zone" {
  type        = string
  description = "Availability zone of the worker subnet"
}

variable "expiration_timestamp" {
  type        = string
  description = "Optional RFC3339 expiration timestamp for automatic cluster deletion"
  default     = null
}

variable "tags" {
  type        = map(string)
  description = "Optional customer AWS tags applied to the AWS resources created for the cluster"
  default     = null
}

variable "oidc_config_id" {
  type        = string
  description = "Optional OIDC config ID (e.g., from OidcConfig resource)"
  default     = ""
}

variable "scheduler_profile" {
  type        = string
  description = "Optional scheduler profile for the hosted cluster"
  default     = null
}

variable "http_proxy" {
  type        = string
  description = "Optional HTTP proxy URL"
  default     = null
}

variable "https_proxy" {
  type        = string
  description = "Optional HTTPS proxy URL"
  default     = null
}

variable "no_proxy" {
  type        = string
  description = "Optional comma-separated list of hosts excluded from proxying"
  default     = null
}

variable "autoscaler_max_node_provision_time" {
  type    = string
  default = null
}

variable "autoscaler_max_nodes_total" {
  type    = number
  default = null
}

variable "autoscaler_max_pod_grace_period" {
  type    = number
  default = null
}

variable "autoscaler_pod_priority_threshold" {
  type    = number
  default = null
}

variable "component_routes" {
  type    = string
  default = null
}

variable "machine_cidr" {
  type        = string
  description = "JSON-encoded machine network entries; changing them requires cluster replacement"
  default     = null
}

variable "pod_cidr" {
  type        = string
  description = "JSON-encoded pod network entries, including hostPrefix; changing them requires cluster replacement"
  default     = null
}

variable "service_cidr" {
  type        = string
  description = "JSON-encoded service network entries; changing them requires cluster replacement"
  default     = null
}
