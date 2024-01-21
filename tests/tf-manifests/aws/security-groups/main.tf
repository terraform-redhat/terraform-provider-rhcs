terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}



module "web_server_sg" {
  count  = var.sg_number
  source = "terraform-aws-modules/security-group/aws//modules/http-80"

  name        = "${var.name_prefix}-${count.index}"
  description = var.description
  vpc_id      = var.vpc_id

  ingress_cidr_blocks = ["11.0.0.0/16"]
}

