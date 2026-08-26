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

resource "rhcs_cluster_hyperfleet" "cluster" {
  name                  = var.cluster_name
  operator_roles_prefix = var.operator_roles_prefix
  aws_subnet_ids        = [var.subnet_id]
  vpc_id                = var.vpc_id
  availability_zones    = [var.availability_zone]
  expiration_timestamp  = var.expiration_timestamp
}
