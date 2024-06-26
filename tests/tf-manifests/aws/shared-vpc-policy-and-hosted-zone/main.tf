provider "aws" {
  region                   = var.region
  shared_credentials_files = var.shared_vpc_aws_shared_credentials_files
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

locals {
  resource_arn_prefix = "arn:${data.aws_partition.current.partition}:ec2:${var.region}:${data.aws_caller_identity.current.account_id}:subnet/"
}

# Private Hosted Zone
resource "aws_route53_zone" "shared_vpc_hosted_zone" {

  name = "${var.cluster_name}.${var.dns_domain_id}"

  vpc {
    vpc_id = var.vpc_id
  }
  lifecycle {
    ignore_changes = [tags]
  }
}

# Shared Role Policy - Trust Policy
resource "aws_iam_role" "shared_vpc_role" {

  name = "${var.cluster_name}-shared-vpc-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          AWS = [
            var.ingress_operator_role_arn,
            var.installer_role_arn
          ]
        }
      }
    ]
  })
  description = "Role that will be assumed from the Target AWS account where the cluster resides"
  lifecycle {
    ignore_changes = [managed_policy_arns]
  }
}

# Shared Role Policy - Policy
resource "aws_iam_policy" "shared_vpc_policy" {

  name = "${var.cluster_name}-shared-vpc-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "route53:ChangeResourceRecordSets",
          "route53:ListHostedZones",
          "route53:ListHostedZonesByName",
          "route53:ListResourceRecordSets",
          "route53:ChangeTagsForResource",
          "route53:GetAccountLimit",
          "route53:GetChange",
          "route53:GetHostedZone",
          "route53:ListTagsForResource",
          "route53:UpdateHostedZoneComment",
          "tag:GetResources",
          "tag:UntagResources"
        ]
        Resource = "*"
      }
    ]
  })
}

# Shared Role Policy - Attachment
resource "aws_iam_role_policy_attachment" "shared_vpc_role_policy_attachment" {

  role       = aws_iam_role.shared_vpc_role.name
  policy_arn = aws_iam_policy.shared_vpc_policy.arn
}

# Resource Share - share subnets to target account
resource "aws_ram_resource_share" "shared_vpc_resource_share" {

  name                      = "${var.cluster_name}-shared-vpc-resource-share"
  allow_external_principals = true
}

resource "aws_ram_principal_association" "shared_vpc_resource_share" {

  principal          = var.cluster_aws_account
  resource_share_arn = aws_ram_resource_share.shared_vpc_resource_share.arn
}

resource "aws_ram_resource_association" "shared_vpc_resource_association" {

  count              = length(var.subnets)
  resource_arn       = "${local.resource_arn_prefix}${var.subnets[count.index]}"
  resource_share_arn = aws_ram_resource_share.shared_vpc_resource_share.arn
}

resource "time_sleep" "shared_resources_propagation" {

  destroy_duration = "20s"
  create_duration  = "20s"

  triggers = {
    shared_vpc_hosted_zone_id = aws_route53_zone.shared_vpc_hosted_zone.id
    shared_vpc_role_arn       = aws_iam_role.shared_vpc_role.arn
    resource_share_arn        = aws_ram_principal_association.shared_vpc_resource_share.resource_share_arn
    policy_arn                = aws_iam_role_policy_attachment.shared_vpc_role_policy_attachment.policy_arn
  }
}

/*
The AZ us-east-1a for VPC-account might not have the same location as us-east-1a for Cluster-account.
For AZs which will be used in cluster configuration, the values should be the ones in Cluster-account.
*/

provider "aws" {
  alias                    = "cluster_account"
  region                   = var.region
  shared_credentials_files = null
}

locals {
  shared_subnets = [for resource_arn in aws_ram_resource_association.shared_vpc_resource_association[*].resource_arn : trimprefix(resource_arn, local.resource_arn_prefix)]
}

data "aws_subnet" "shared_subnets" {
  provider   = aws.cluster_account
  for_each   = toset(local.shared_subnets)
  vpc_id     = var.vpc_id
  id         = each.value
  depends_on = [time_sleep.shared_resources_propagation]
}
