terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    red-hat-cloud-services = {
      version = ">= 0.0.1"
      source  = "terraform.local/local/red-hat-cloud-services"
    }
  }
}

provider "red-hat-cloud-services" {
  token = var.token
  url   = var.url
}

data "ocm_policies" "all_policies" {}

module "create_account_roles" {
  source  = "terraform-redhat/rosa-sts/aws"
  version = "0.0.8"

  create_operator_roles = false
  create_oidc_provider  = false
  create_account_roles  = true

  account_role_prefix    = var.account_role_prefix
  ocm_environment        = var.ocm_environment
  rosa_openshift_version = "${split(".", var.openshift_version)[0]}.${split(".", var.openshift_version)[1]}"
  account_role_policies  = data.ocm_policies.all_policies.account_role_policies
  operator_role_policies = data.ocm_policies.all_policies.operator_role_policies
}
