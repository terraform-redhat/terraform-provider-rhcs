#
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    ocm = {
      version = ">= 0.0.1"
      source  = "terraform.local/local/ocm"
    }
  }
}

provider "ocm" {
  token = var.token
  url = var.url
}

data "ocm_policies" "all_policies"{}

module "create_account_roles"{
  source = "terraform-redhat/rosa-sts/aws"
  version = "0.0.4"

  create_operator_roles = false
  create_oidc_provider = false
  create_account_roles = true

  account_role_prefix =  var.account_role_prefix
  ocm_environment =  var.ocm_environment
  rosa_openshift_version=  var.openshift_version
  account_role_policies=data.ocm_policies.all_policies.account_role_policies
  operator_role_policies=data.ocm_policies.all_policies.operator_role_policies
}
