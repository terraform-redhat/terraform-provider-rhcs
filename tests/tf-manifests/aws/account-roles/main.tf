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
  token = var.token
  url   = var.url
}

locals {
  major_version = "${split(".", var.openshift_version)[0]}.${split(".", var.openshift_version)[1]}"
  versionfilter = var.openshift_version == null ? "" : " and raw_id like '%${local.major_version}%'"
}

data "rhcs_policies" "all_policies" {}

data "rhcs_versions" "all" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
  order  = "id"
}

module "create_account_roles" {
  source  = "terraform-redhat/rosa-sts/aws"
  version = ">= 0.0.14"

  create_operator_roles  = false
  create_oidc_provider   = false
  create_account_roles   = true
  all_versions           = data.rhcs_versions.all
  account_role_prefix    = var.account_role_prefix
  ocm_environment        = var.rhcs_environment
  rosa_openshift_version = local.major_version
  account_role_policies  = data.rhcs_policies.all_policies.account_role_policies
  operator_role_policies = data.rhcs_policies.all_policies.operator_role_policies
}
