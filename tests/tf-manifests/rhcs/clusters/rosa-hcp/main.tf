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

data "rhcs_versions" "version" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'"
  order  = "id"
}

locals {
  version = var.openshift_version != null ? var.openshift_version : data.rhcs_versions.version.items[0].name

  creatorProps = {
      rosa_creator_arn = data.aws_caller_identity.current.arn
    }
  properties = var.include_creator_property ? merge(local.creatorProps, var.custom_properties) : var.custom_properties
}

locals {
  account_role_path = coalesce(var.path, "/")

  sts_roles = {
    role_arn         = var.installer_role != null ? var.installer_role : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.account_role_path}${var.account_role_prefix}-HCP-ROSA-Installer-Role",
    support_role_arn = var.support_role != null ? var.support_role : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.account_role_path}${var.account_role_prefix}-HCP-ROSA-Support-Role",
    instance_iam_roles = {
      worker_role_arn = var.worker_role != null ? var.worker_role : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.account_role_path}${var.account_role_prefix}-HCP-ROSA-Worker-Role"
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
  aws_account_id         = var.aws_account_id != null ? var.aws_account_id : data.aws_caller_identity.current.account_id
  aws_billing_account_id = var.aws_billing_account_id != null ? var.aws_billing_account_id : data.aws_caller_identity.current.account_id
  availability_zones     = var.aws_availability_zones
  properties = local.properties
  sts      = local.sts_roles
  replicas = var.replicas
  proxy    = var.proxy
  aws_subnet_ids               = var.aws_subnet_ids
  private                      = var.private
  compute_machine_type         = var.compute_machine_type
  etcd_encryption              = var.etcd_encryption
  etcd_kms_key_arn             = var.etcd_kms_key_arn != null ? var.etcd_kms_key_arn : var.kms_key_arn
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
