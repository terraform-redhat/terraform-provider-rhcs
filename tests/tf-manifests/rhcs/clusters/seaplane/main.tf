# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

data "aws_caller_identity" "current" {}

provider "rhcs" {
  seaplane_url   = var.seaplane_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
}

resource "rhcs_cluster_seaplane" "cluster" {
  name                  = var.cluster_name
  operator_roles_prefix = var.operator_roles_prefix
  version               = var.release_image
  network = var.machine_cidr == null && var.pod_cidr == null && var.service_cidr == null ? null : {
    machine_cidr = var.machine_cidr
    pod_cidr     = var.pod_cidr
    service_cidr = var.service_cidr
  }
  aws = {
    aws_subnet_ids     = [var.subnet_id]
    vpc_id             = var.vpc_id
    availability_zones = [var.availability_zone]
  }
  expiration_timestamp               = var.expiration_timestamp
  oidc_config_id                     = var.oidc_config_id != "" ? var.oidc_config_id : null
  scheduler_profile                  = var.scheduler_profile
  tags                               = var.tags
  http_proxy                         = var.http_proxy
  https_proxy                        = var.https_proxy
  no_proxy                           = var.no_proxy
  autoscaler_max_node_provision_time = var.autoscaler_max_node_provision_time
  autoscaler_max_nodes_total         = var.autoscaler_max_nodes_total
  autoscaler_max_pod_grace_period    = var.autoscaler_max_pod_grace_period
  autoscaler_pod_priority_threshold  = var.autoscaler_pod_priority_threshold
  component_routes                   = var.component_routes
}
