
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
  path    = coalesce(var.path, "/")
  managed = var.oidc_config == null || var.oidc_config == "" || var.oidc_config == "managed"
}

# Create OIDC config and provider in OCM
module "oidc_config_and_provider" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/oidc-config-and-provider"
  version = ">=1.6.3"

  installer_role_arn = local.managed ? null : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Installer-Role"
  managed            = local.managed
  tags               = var.tags
}

# Create operator roles on AWS
module "operator_roles" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/operator-roles"
  version = ">=1.6.3"

  account_role_prefix  = var.account_role_prefix
  oidc_endpoint_url    = module.oidc_config_and_provider.oidc_endpoint_url
  operator_role_prefix = var.operator_role_prefix
  path                 = local.path
  permissions_boundary = var.permissions_boundary
  tags                 = var.tags
}
