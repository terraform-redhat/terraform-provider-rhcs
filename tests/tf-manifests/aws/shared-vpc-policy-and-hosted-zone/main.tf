provider "aws" {
  region                   = var.region
  shared_credentials_files = var.shared_vpc_aws_shared_credentials_files
}

module "shared_vpc_policy_and_hosted_zone" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/shared-vpc-policy-and-hosted-zone"
  version = ">=1.6.3"

  cluster_name              = var.domain_prefix == null ? var.cluster_name : var.domain_prefix
  hosted_zone_base_domain   = var.dns_domain_id
  ingress_operator_role_arn = var.ingress_operator_role_arn
  installer_role_arn        = var.installer_role_arn
  name_prefix               = var.cluster_name
  subnets                   = var.subnets
  target_aws_account        = var.cluster_aws_account
  vpc_id                    = var.vpc_id
}

provider "aws" {
  alias                    = "cluster_account"
  region                   = var.region
  shared_credentials_files = null
}

data "aws_subnet" "shared_subnets" {
  provider   = aws.cluster_account
  for_each   = toset(module.shared_vpc_policy_and_hosted_zone.shared_subnets)
  vpc_id     = var.vpc_id
  id         = each.value
  depends_on = [module.shared_vpc_policy_and_hosted_zone]
}