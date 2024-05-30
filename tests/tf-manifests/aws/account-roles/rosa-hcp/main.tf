terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.38.0"
    }
    rhcs = {
      version = ">= 1.5.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

locals {
  major_version = "${split(".", var.openshift_version)[0]}.${split(".", var.openshift_version)[1]}"
  versionfilter = var.openshift_version == null ? "" : " and raw_id like '%${local.major_version}%'"
}

data "rhcs_versions" "all_versions" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
  order  = "id"
}


###########################################################################
## Below is taken from rhcs-hcp module `account-iam-resources`
###########################################################################

locals {
  path = coalesce(var.path, "/")
  account_roles_properties = [
    {
      role_name            = "HCP-ROSA-Installer"
      role_type            = "installer"
      policy_details       = data.rhcs_hcp_policies.all_policies.account_role_policies["sts_hcp_installer_permission_policy"]
      principal_type       = "AWS"
      principal_identifier = "arn:aws:iam::${data.rhcs_info.current.ocm_aws_account_id}:role/RH-Managed-OpenShift-Installer"
    },
    {
      role_name            = "HCP-ROSA-Support"
      role_type            = "support"
      policy_details       = data.rhcs_hcp_policies.all_policies.account_role_policies["sts_hcp_support_permission_policy"]
      principal_type       = "AWS"
      principal_identifier = "arn:aws:iam::${data.rhcs_info.current.ocm_aws_account_id}:role/RH-Technical-Support-Access"
    },
    {
      role_name            = "HCP-ROSA-Worker"
      role_type            = "instance_worker"
      policy_details       = data.rhcs_hcp_policies.all_policies.account_role_policies["sts_hcp_instance_worker_permission_policy"]
      principal_type       = "Service"
      principal_identifier = "ec2.amazonaws.com"
    },
  ]
  account_roles_count = length(local.account_roles_properties)
  patch_version_list  = [for s in data.rhcs_versions.all_versions.items : s.name]
  minor_version_list = length(local.patch_version_list) > 0 ? (
    distinct([for s in local.patch_version_list : format("%s.%s", split(".", s)[0], split(".", s)[1])])
    ) : (
    []
  )
  account_role_prefix_valid = var.account_role_prefix != null ? (
    var.account_role_prefix
    ) : (
    "account-role-${random_string.default_random[0].result}"
  )
}

data "aws_iam_policy_document" "custom_trust_policy" {
  count = local.account_roles_count

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = local.account_roles_properties[count.index].principal_type
      identifiers = [local.account_roles_properties[count.index].principal_identifier]
    }
  }
}

module "account_iam_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">=5.34.0"
  count   = local.account_roles_count

  create_role = true

  role_name = "${local.account_role_prefix_valid}-${local.account_roles_properties[count.index].role_name}-Role"

  role_path                     = local.path
  role_permissions_boundary_arn = var.permissions_boundary

  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.custom_trust_policy[count.index].json

  custom_role_policy_arns = [
    "${local.account_roles_properties[count.index].policy_details}"
  ]

  tags = merge(var.tags, {
    rosa_hcp_policies     = true
    red-hat-managed       = true
    rosa_role_prefix      = "${local.account_role_prefix_valid}"
    rosa_role_type        = "${local.account_roles_properties[count.index].role_type}"
    rosa_managed_policies = true
  })
}

data "rhcs_hcp_policies" "all_policies" {}

resource "random_string" "default_random" {
  count = var.account_role_prefix != null ? 0 : 1

  length  = 4
  special = false
  upper   = false
}

data "rhcs_info" "current" {}

resource "time_sleep" "wait_10_seconds" {
  destroy_duration = "10s"
  create_duration  = "10s"
  triggers = {
    account_iam_role_name = jsonencode([for value in module.account_iam_role : value.iam_role_name])
    account_roles_arn     = jsonencode({ for idx, value in module.account_iam_role : local.account_roles_properties[idx].role_name => value.iam_role_arn })
    account_role_prefix   = local.account_role_prefix_valid
    path                  = var.path
  }
}
