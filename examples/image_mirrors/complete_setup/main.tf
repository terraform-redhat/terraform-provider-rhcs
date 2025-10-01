terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform-redhat/rhcs"
    }
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
}

provider "rhcs" {
}

provider "aws" {
  region = var.aws_region
}

locals {
  sts_roles = {
    role_arn         = var.installer_role_arn
    support_role_arn = var.support_role_arn
    instance_iam_roles = {
      worker_role_arn = var.worker_role_arn
    }
    operator_role_prefix = var.operator_role_prefix
    oidc_config_id       = var.oidc_config_id
  }
}

# Create ROSA HCP cluster
resource "rhcs_cluster_rosa_hcp" "cluster" {
  name                   = var.cluster_name
  cloud_region           = var.aws_region
  aws_account_id         = var.aws_account_id
  aws_billing_account_id = var.aws_billing_account_id
  aws_subnet_ids         = var.subnet_ids
  availability_zones     = var.availability_zones
  replicas               = var.replicas
  version                = var.openshift_version

  properties = var.cluster_properties
  sts        = local.sts_roles

  wait_for_create_complete            = true
  wait_for_std_compute_nodes_complete = true

  tags = var.cluster_tags
}

# Create image mirrors after cluster is ready
resource "rhcs_image_mirror" "cluster_mirrors" {
  for_each = var.image_mirrors

  cluster_id = rhcs_cluster_rosa_hcp.cluster.id
  source     = each.key
  mirrors    = each.value
  type       = "digest"

  depends_on = [rhcs_cluster_rosa_hcp.cluster]
}

# Data source to retrieve all configured mirrors
data "rhcs_image_mirrors" "all_mirrors" {
  cluster_id = rhcs_cluster_rosa_hcp.cluster.id

  depends_on = [rhcs_image_mirror.cluster_mirrors]
}