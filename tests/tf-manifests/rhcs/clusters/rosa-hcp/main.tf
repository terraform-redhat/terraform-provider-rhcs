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
  name                         = var.cluster_name
  version                      = local.version
  channel_group                = var.channel_group
  cloud_region                 = var.aws_region
  aws_account_id               = var.aws_account_id != null ? var.aws_account_id : data.aws_caller_identity.current.account_id
  aws_billing_account_id       = var.aws_billing_account_id != null ? var.aws_billing_account_id : data.aws_caller_identity.current.account_id
  availability_zones           = var.aws_availability_zones
  properties                   = local.properties
  sts                          = local.sts_roles
  replicas                     = var.replicas
  proxy                        = var.proxy
  aws_subnet_ids               = var.aws_subnet_ids
  private                      = var.private
  compute_machine_type         = var.compute_machine_type
  ec2_metadata_http_tokens     = var.ec2_metadata_http_tokens
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
  aws_additional_compute_security_group_ids       = var.additional_compute_security_groups
  wait_for_create_complete                        = var.wait_for_cluster
  wait_for_std_compute_nodes_complete             = var.wait_for_cluster
  disable_waiting_in_destroy                      = var.disable_waiting_in_destroy
  registry_config                                 = var.registry_config
  worker_disk_size                                = var.worker_disk_size
  external_auth_providers_enabled                 = var.external_auth_providers_enabled
  log_forwarders_at_cluster_creation                     = var.log_forwarders_at_cluster_creation
}

resource "rhcs_cluster_wait" "rosa_cluster" { # id: 71869
  count   = var.disable_cluster_waiter || !var.wait_for_cluster ? 0 : 1
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  timeout = 60 # in minutes
}

// Data source to query all log forwarders (both Day 1 and Day 2)
data "rhcs_log_forwarders" "all" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
}

// Output full details of all log forwarders (from data source)
output "log_forwarders" {
  description = "Full details of all log forwarders"
  value       = data.rhcs_log_forwarders.all.items
}

resource "rhcs_hcp_default_ingress" "current" {
  count            = var.full_resources ? 1 : 0
  cluster          = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  listening_method = "internal"
}
locals {
  default_mp_name = length(rhcs_cluster_rosa_hcp.rosa_hcp_cluster.availability_zones) == 1 ? "workers" : "workers-0"
}
data "rhcs_hcp_machine_pool" "default_machine_pool" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  name    = local.default_mp_name
}
// machinepool resource should wait for the cluster creation finished
resource "rhcs_hcp_machine_pool" "mp" {
  count   = var.full_resources ? 1 : 0
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  aws_node_pool = {
    instance_type = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.compute_machine_type,
  }
  autoscaling = {
    enabled = false
  }
  name        = "full-resource"
  replicas    = 0
  auto_repair = false
  subnet_id   = data.rhcs_hcp_machine_pool.default_machine_pool.subnet_id
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

// idp resource should wait for the cluster creation finished
resource "rhcs_identity_provider" "htpasswd_idp" {
  count          = var.full_resources ? 1 : 0
  cluster        = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  name           = "full-resource"
  mapping_method = "claim"
  htpasswd = {
    users = [{
      username = "full-resource"
      password = random_password.password.result
    }]
  }
}

// kubeletconfig will be created after cluster created
resource "rhcs_kubeletconfig" "kubeletconfig" {
  count          = var.full_resources ? 1 : 0
  cluster        = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  pod_pids_limit = 4097
}

locals {
  defaultSpec = jsonencode(
    {
      "profile" : [
        {
          "data" : "[main]\nsummary=Custom OpenShift profile\ninclude=openshift-node\n\n[sysctl]\nvm.dirty_ratio=\"65\"\n",
          "name" : "tuned-profile"
        }
      ],
      "recommend" : [
        {
          "priority" : 10,
          "profile" : "tuned-profile"
        }
      ]
    }
  )
}

resource "rhcs_tuning_config" "tcs" {
  count   = var.full_resources ? 1 : 0
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  name    = "full-resource"
  spec    = local.defaultSpec
}
