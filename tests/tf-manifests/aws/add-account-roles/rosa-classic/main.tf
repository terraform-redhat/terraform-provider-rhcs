terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.1.0"
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
  path               = coalesce(var.path, "/")
  major_version      = "${split(".", var.openshift_version)[0]}.${split(".", var.openshift_version)[1]}"
  versionfilter      = var.openshift_version == null ? "" : " and raw_id like '%${local.major_version}%'"
  installer_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Installer-Role"
  aws_account_id     = data.aws_caller_identity.current.account_id
}

data "rhcs_policies" "all_policies" {}

data "rhcs_versions" "all" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
  order  = "id"
}

module "create_account_roles" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/account-iam-resources"
  version = ">=1.5.0"

  account_role_prefix = var.account_role_prefix
  openshift_version   = var.openshift_version
  path                = local.path
}
