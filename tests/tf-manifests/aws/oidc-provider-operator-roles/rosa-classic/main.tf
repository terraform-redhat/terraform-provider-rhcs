
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

provider "rhcs" {
}

provider "aws" {
  region = var.aws_region
}
data "aws_caller_identity" "current" {
}
data "aws_partition" "current" {
}

locals {
  path                      = coalesce(var.path, "/")
  ingress_role_name         = substr("${var.operator_role_prefix}-openshift-ingress-operator-cloud-credentials", 0, 64)
  ingress_operator_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${local.ingress_role_name}"
  managed                   = var.oidc_config == "managed"
}

# Create OIDC config and provider in OCM
module "oidc_config_and_provider" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/oidc-config-and-provider"
  version = ">=1.5.0"

  installer_role_arn = local.managed ? null : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Installer-Role"
  managed            = local.managed
}

# Create operator roles on AWS
module "operator_roles" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/operator-roles"
  version = ">=1.5.0"

  account_role_prefix  = var.account_role_prefix
  oidc_endpoint_url    = module.oidc_config_and_provider.oidc_endpoint_url
  operator_role_prefix = var.operator_role_prefix
  path                 = local.path
}
