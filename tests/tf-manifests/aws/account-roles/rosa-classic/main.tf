terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.6.3-0"
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
  source  = "terraform-redhat/rosa-classic/rhcs//modules/account-iam-resources"
  version = ">=1.6.3-0"

  account_role_prefix  = var.account_role_prefix
  openshift_version    = var.openshift_version
  path                 = local.path
  permissions_boundary = var.permissions_boundary
  tags                 = var.tags
}

module "rosa-classic_operator-policies" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/operator-policies"
  version = ">=1.6.3-0"

  account_role_prefix = module.create_account_roles.account_role_prefix
  openshift_version   = var.openshift_version
  path                = local.path
  shared_vpc_role_arn = var.shared_vpc_role_arn
  tags                = var.tags
}

resource "random_string" "random_suffix" {
  count = var.account_role_prefix == "" || var.account_role_prefix == null ? 1 : 0

  length  = 4
  special = false
  upper   = false
}
