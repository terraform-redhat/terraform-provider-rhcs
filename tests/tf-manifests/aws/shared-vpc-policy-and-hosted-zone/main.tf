provider "aws" {
  region                   = var.region
  shared_credentials_files = var.shared_vpc_aws_shared_credentials_files
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

locals {
  prefix = var.domain_prefix == null ? var.cluster_name : var.domain_prefix
}

module "shared_vpc_policy_and_hosted_zone" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/shared-vpc-policy-and-hosted-zone"
  version = ">=1.6.3"

  cluster_name              = var.cluster_name
  hosted_zone_base_domain   = "${local.prefix}.${var.dns_domain_id}"
  ingress_operator_role_arn = var.ingress_operator_role_arn
  installer_role_arn        = var.installer_role_arn
  name_prefix               = var.cluster_name
  subnets                   = var.subnets
  target_aws_account        = var.cluster_aws_account
  vpc_id                    = var.vpc_id
}
