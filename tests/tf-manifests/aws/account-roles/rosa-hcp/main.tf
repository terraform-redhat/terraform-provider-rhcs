terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0, != 5.71.0"
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
  source  = "terraform-redhat/rosa-hcp/rhcs//modules/account-iam-resources"
  version = ">=1.6.3"

  account_role_prefix  = var.account_role_prefix
  path                 = local.path
  permissions_boundary = var.permissions_boundary
  tags                 = var.tags
}
