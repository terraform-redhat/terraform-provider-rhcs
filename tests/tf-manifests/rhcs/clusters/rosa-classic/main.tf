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
}

locals {
  aws_subnet_ids = var.aws_subnet_ids
}

locals {
  path = coalesce(var.path, "/")
  sts_roles = {
    role_arn         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Worker-Role"
    },
    operator_role_prefix = var.operator_role_prefix
    oidc_config_id       = var.oidc_config_id
  }
}

data "aws_caller_identity" "current" {
}

resource "rhcs_cluster_rosa_classic" "rosa_sts_cluster" {
  name               = var.cluster_name
  version            = local.version
  channel_group      = var.channel_group
  cloud_region       = var.aws_region
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = var.aws_availability_zones
  multi_az           = var.multi_az
  properties = merge(
    {
      rosa_creator_arn = data.aws_caller_identity.current.arn
    },
    var.custom_properties
  )
  sts                                             = local.sts_roles
  replicas                                        = var.replicas
  proxy                                           = var.proxy
  autoscaling_enabled                             = var.autoscaling.autoscaling_enabled
  min_replicas                                    = var.autoscaling.min_replicas
  max_replicas                                    = var.autoscaling.max_replicas
  ec2_metadata_http_tokens                        = var.ec2_metadata_http_tokens
  aws_private_link                                = var.private_link
  private                                         = var.private
  aws_subnet_ids                                  = local.aws_subnet_ids
  compute_machine_type                            = var.compute_machine_type
  default_mp_labels                               = var.default_mp_labels
  disable_scp_checks                              = var.disable_scp_checks
  disable_workload_monitoring                     = var.disable_workload_monitoring
  etcd_encryption                                 = var.etcd_encryption
  fips                                            = var.fips
  host_prefix                                     = var.host_prefix
  kms_key_arn                                     = var.kms_key_arn
  machine_cidr                                    = var.machine_cidr
  service_cidr                                    = var.service_cidr
  pod_cidr                                        = var.pod_cidr
  tags                                            = var.tags
  admin_credentials                               = var.admin_credentials
  worker_disk_size                                = var.worker_disk_size
  aws_additional_compute_security_group_ids       = var.additional_compute_security_groups
  aws_additional_infra_security_group_ids         = var.additional_infra_security_groups
  aws_additional_control_plane_security_group_ids = var.additional_control_plane_security_groups
  destroy_timeout                                 = 120
  upgrade_acknowledgements_for                    = var.upgrade_acknowledgements_for
  base_dns_domain                                 = var.base_dns_domain
  private_hosted_zone                             = var.private_hosted_zone
  lifecycle {
    ignore_changes = [availability_zones]
  }
  wait_for_create_complete = true
}

resource "rhcs_cluster_wait" "rosa_cluster" { # id: 71869
  count   = var.deactivate_cluster_waiter ? 0 : 1
  cluster = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  timeout = 120
}
