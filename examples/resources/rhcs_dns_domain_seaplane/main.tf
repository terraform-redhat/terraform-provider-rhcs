data "aws_caller_identity" "current" {}

provider "rhcs" {
  seaplane_url   = var.seaplane_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
  aws_region     = var.aws_region
}

resource "rhcs_dns_domain_seaplane" "domain" {
  cluster_arch = "hcp"
}

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API endpoint."
}

variable "aws_region" {
  type        = string
  description = "AWS region used to sign Seaplane Platform API requests."
}
