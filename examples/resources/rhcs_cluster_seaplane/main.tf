data "aws_caller_identity" "current" {}

data "aws_subnet" "worker" {
  id = var.subnet_id
}

provider "rhcs" {
  seaplane_url   = var.seaplane_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
}

resource "rhcs_cluster_seaplane" "cluster" {
  name                  = var.cluster_name
  operator_roles_prefix = var.operator_roles_prefix
  aws = {
    aws_subnet_ids                = [data.aws_subnet.worker.id]
    vpc_id                        = data.aws_subnet.worker.vpc_id
    availability_zones            = [data.aws_subnet.worker.availability_zone]
    additional_allowed_principals = ["arn:aws:iam::123456789012:role/example-consumer"]
  }
  version = var.release_image
}
