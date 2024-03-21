terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  url = var.url
}

provider "aws" {
  region = var.aws_region
}

locals {
  versionfilter = var.openshift_version == null ? "" : " and id like '%${var.openshift_version}%'"
}

data "rhcs_versions" "version" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
  order  = "id"
}

locals {
  version = data.rhcs_versions.version.items[0].name
}

locals {
  account_role_path = coalesce(var.path, "/")

  sts_roles = {
    role_arn         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.account_role_path}${var.account_role_prefix}-HCP-ROSA-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.account_role_path}${var.account_role_prefix}-HCP-ROSA-Support-Role",
    instance_iam_roles = {
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.account_role_path}${var.account_role_prefix}-HCP-ROSA-Worker-Role"
    }
    operator_role_prefix = var.operator_role_prefix
    oidc_config_id       = var.oidc_config_id
  }
}

data "aws_caller_identity" "current" {
}

resource "rhcs_cluster_rosa_hcp" "rosa_hcp_cluster" {
  name                   = var.cluster_name
  version                = local.version
  channel_group          = var.channel_group
  cloud_region           = var.aws_region
  aws_account_id         = data.aws_caller_identity.current.account_id
  aws_billing_account_id = data.aws_caller_identity.current.account_id
  availability_zones     = var.aws_availability_zones
  properties = merge(
    {
      rosa_creator_arn = data.aws_caller_identity.current.arn
    },
    var.custom_properties
  )
  sts      = local.sts_roles
  replicas = var.replicas
  proxy    = var.proxy
  # autoscaling_enabled                             = var.autoscaling.autoscaling_enabled
  # min_replicas                                    = var.autoscaling.min_replicas
  # max_replicas                                    = var.autoscaling.max_replicas
  aws_subnet_ids               = var.aws_subnet_ids
  private                      = var.private
  compute_machine_type         = var.compute_machine_type
  etcd_encryption              = var.etcd_encryption
  etcd_kms_key_arn             = var.kms_key_arn
  kms_key_arn                  = var.kms_key_arn
  host_prefix                  = var.host_prefix
  machine_cidr                 = var.machine_cidr
  service_cidr                 = var.service_cidr
  pod_cidr                     = var.pod_cidr
  tags                         = var.tags
  destroy_timeout              = 60
  upgrade_acknowledgements_for = var.upgrade_acknowledgements_for
  lifecycle {
    ignore_changes = [availability_zones]
  }
  wait_for_create_complete   = true
  disable_waiting_in_destroy = false
}

resource "rhcs_cluster_wait" "rosa_cluster" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  timeout = 60 # in minutes
}
