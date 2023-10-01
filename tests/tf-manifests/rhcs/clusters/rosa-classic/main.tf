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
  token = var.token
  url   = var.url
}
provider "aws" {
  region = var.aws_region
}
locals {
  versionfilter = var.openshift_version == null ? "" : " and id like '%${var.openshift_version}%'"
}

data "rhcs_versions" "version" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
  order = "id"
}
locals {
  version = data.rhcs_versions.version.items[0].name
}

locals {
  aws_subnet_ids = var.aws_subnet_ids
}

locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Worker-Role"
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
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts                         = local.sts_roles
  replicas                    = var.replicas
  proxy                       = var.proxy
  autoscaling_enabled         = var.autoscaling.autoscaling_enabled
  min_replicas                = var.autoscaling.min_replicas
  max_replicas                = var.autoscaling.max_replicas
  ec2_metadata_http_tokens    = var.aws_http_tokens_state
  aws_private_link            = var.private_link
  aws_subnet_ids              = local.aws_subnet_ids
  compute_machine_type        = var.compute_machine_type
  default_mp_labels           = var.default_mp_labels
  disable_scp_checks          = var.disable_scp_checks
  disable_workload_monitoring = var.disable_workload_monitoring
  etcd_encryption             = var.etcd_encryption
  fips                        = var.fips
  host_prefix                 = var.host_prefix
  kms_key_arn                 = var.kms_key_arn
  machine_cidr                = var.machine_cidr
  service_cidr                = var.service_cidr
  pod_cidr                    = var.pod_cidr
  tags                        = var.tags
  destroy_timeout             = 120
  lifecycle {
    ignore_changes = [availability_zones]
  }
}

resource "rhcs_cluster_wait" "rosa_cluster" {
  cluster = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  timeout = 120
}
