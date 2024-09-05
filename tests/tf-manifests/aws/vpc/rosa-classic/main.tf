terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
  }
}



provider "aws" {
  region                   = var.aws_region
  shared_credentials_files = var.aws_shared_credentials_files
}

module "vpc" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/vpc"
  version = ">=1.6.3"

  vpc_cidr                 = var.vpc_cidr
  name_prefix              = var.name_prefix
  availability_zones       = var.availability_zones
  availability_zones_count = var.availability_zones_count
  tags                     = var.tags
}
