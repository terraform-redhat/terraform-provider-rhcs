# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      source  = "terraform.local/local/rhcs"
      version = ">= 1.0.1"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

data "aws_caller_identity" "current" {}

provider "rhcs" {
  seaplane_url   = var.seaplane_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
  aws_region     = var.aws_region
}

resource "rhcs_dns_domain_seaplane" "dns_domain" {
  name = var.name
}
