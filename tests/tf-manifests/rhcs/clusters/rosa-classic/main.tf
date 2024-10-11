terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0, != 5.71.0"
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
  domain_prefix                                   = var.domain_prefix
  private_hosted_zone                             = var.private_hosted_zone
  lifecycle {
    ignore_changes = [availability_zones]
  }
  wait_for_create_complete   = var.wait_for_cluster
  disable_waiting_in_destroy = var.disable_waiting_in_destroy
}

resource "rhcs_cluster_wait" "rosa_cluster" { # id: 71869
  count   = var.disable_cluster_waiter || !var.wait_for_cluster ? 0 : 1
  cluster = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  timeout = 120
}

// autoscaler resource should wait for cluster creation finished
resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
  count                = var.full_resources ? 1 : 0
  cluster              = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  max_pod_grace_period = 1000
}

// machinepool resource should wait for the cluster creation finished
resource "rhcs_machine_pool" "mp" {
  count                   = var.full_resources ? 1 : 0
  cluster                 = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  machine_type            = rhcs_cluster_rosa_classic.rosa_sts_cluster.compute_machine_type
  name                    = "full-resource"
  replicas                = 3
  multi_availability_zone = rhcs_cluster_rosa_classic.rosa_sts_cluster.multi_az
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

// idp resource should wait for the cluster creation finished
resource "rhcs_identity_provider" "htpasswd_idp" {
  count          = var.full_resources ? 1 : 0
  cluster        = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  name           = "full-resource"
  mapping_method = "claim"
  htpasswd = {
    users = [{
      username = "full-resource"
      password = random_password.password.result
    }]
  }
}

// ingress resource should be changed after cluster created
resource "rhcs_default_ingress" "default_ingress" {
  count               = var.full_resources ? 1 : 0
  cluster             = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  excluded_namespaces = ["full-resource"]
}

// kubeletconfig will be created after cluster created
resource "rhcs_kubeletconfig" "kubeletconfig" {
  count          = var.full_resources ? 1 : 0
  cluster        = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  pod_pids_limit = 4097
}

// user will be created after cluster created
resource "rhcs_group_membership" "group_membership" {
  count   = var.full_resources ? 1 : 0
  cluster = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  group   = "cluster-admins"
  user    = "full-resource"

}