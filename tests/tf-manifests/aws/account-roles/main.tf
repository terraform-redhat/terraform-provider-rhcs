terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 0.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  token = var.token
  url = var.url
}

data "rhcs_policies" "all_policies"{}

module "create_account_roles"{
  source = "terraform-redhat/rosa-sts/aws"
  version = "0.0.9"

  create_operator_roles = false
  create_oidc_provider = false
  create_account_roles = true

  account_role_prefix    = var.account_role_prefix
  ocm_environment        = var.rhcs_environment
  rosa_openshift_version = "${split(".", var.openshift_version***REMOVED***[0]}.${split(".", var.openshift_version***REMOVED***[1]}"
  account_role_policies  = data.rhcs_policies.all_policies.account_role_policies
  operator_role_policies = data.rhcs_policies.all_policies.operator_role_policies
}