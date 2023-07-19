terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
  }
}



provider "aws" {
  region  = var.aws_region
}
locals {
  private_cidr_map = {
    "10.0.0.0/16" = ["10.0.1.0/24", "10.0.2.0/24","10.0.3.0/24"],
    "11.0.0.0/16" = ["11.0.1.0/24", "11.0.2.0/24","11.0.3.0/24"],
    "12.0.0.0/16" = ["12.0.1.0/24", "12.0.2.0/24","12.0.3.0/24"],
  }
  public_cidr_map = {
    "10.0.0.0/16" = ["10.0.101.0/24", "10.0.102.0/24","10.0.103.0/24"],
    "11.0.0.0/16" = ["11.0.101.0/24", "11.0.102.0/24","11.0.103.0/24"],
    "12.0.0.0/16" = ["12.0.101.0/24", "12.0.102.0/24","12.0.103.0/24"],
  }
}

locals {
  private_subnets = var.multi_az?local.private_cidr_map[var.vpc_cidr]:[local.private_cidr_map[var.vpc_cidr][0]]
  public_subnets = var.multi_az?local.public_cidr_map[var.vpc_cidr]:[local.public_cidr_map[var.vpc_cidr][0]]
}

locals {
  azs =var.multi_az?var.az_ids[var.aws_region]:[var.az_ids[var.aws_region][0]]
}
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = ">= 4.0.0"

  name = "${var.cluster_name}-vpc"
  cidr = var.vpc_cidr

  # azs             = local.azs
  azs = var.test_zones
  private_subnets = local.private_subnets
  public_subnets  = local.public_subnets

  enable_nat_gateway   = true
  single_nat_gateway   = var.multi_az
  enable_dns_hostnames = true
  enable_dns_support   = true
}

