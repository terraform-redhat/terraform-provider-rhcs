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
  url = var.url
}

locals {
  versionfilter = var.openshift_version == null ? "" : " and id like '%${var.openshift_version}%'"
}

data "rhcs_policies" "all_policies"{}

data "rhcs_versions" "all" {
    search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
    order = "id"
}

module "create_account_roles"{
  source = "terraform-redhat/rosa-sts/aws"
  version = ">= 0.0.12"

  create_operator_roles = false
  create_oidc_provider = false
  create_account_roles = true
  all_versions           = data.rhcs_versions.all

  account_role_prefix    = var.account_role_prefix
  ocm_environment        = var.rhcs_environment
  # rosa_openshift_version = "${split(".", var.openshift_version***REMOVED***[0]}.${split(".", var.openshift_version***REMOVED***[1]}"
  rosa_openshift_version= var.openshift_version
  account_role_policies  = data.rhcs_policies.all_policies.account_role_policies
  operator_role_policies = data.rhcs_policies.all_policies.operator_role_policies
}