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
  hyperfleet_url = var.hyperfleet_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
}

resource "rhcs_nodepool_hyperfleet" "nodepool" {
  cluster     = var.cluster_id
  name        = var.name
  subnet_id   = var.subnet_id
  auto_repair = var.auto_repair
  replicas      = var.replicas

  aws_node_pool = {
    instance_type = var.instance_type
    disk_size     = var.disk_size
    tags          = var.tags
  }

  labels = var.labels
}
