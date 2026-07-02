# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.6.3"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

data "aws_caller_identity" "current" {
}

data "aws_partition" "current" {
}

locals {
  path           = coalesce(var.path, "/")
  aws_account_id = data.aws_caller_identity.current.account_id
}

module "create_account_roles" {
  source = "git::https://github.com/terraform-redhat/terraform-rhcs-rosa-hcp//modules/account-iam-resources?ref=main"

  account_role_prefix      = var.account_role_prefix
  path                     = local.path
  permissions_boundary     = var.permissions_boundary
  tags                     = var.tags
  trust_policy_external_id = var.trust_policy_external_id
}
