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
  ignore_tags {
    key_prefixes = ["kubernetes.io/cluster/"]
  }
}

resource "aws_ec2_tag" "tagging" {
  for_each    = toset(var.ids)
  resource_id = each.value
  key         = var.key
  value       = var.value
}