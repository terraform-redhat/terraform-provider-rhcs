#
# Copyright (c) 2023 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform-redhat/rhcs"
    }
  }
}
provider "rhcs" {
  token = var.token
  url   = var.url
}

locals {
  installer_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Installer-Role"
}

# Create unmanaged OIDC config
module "oidc_config" {
  token                = var.token
  url                  = var.url
  source               = "../oidc_provider"
  managed              = false
  installer_role_arn   = local.installer_role_arn
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix  = var.account_role_prefix
  cloud_region         = var.cloud_region
  tags                 = var.tags
  path                 = var.path
}

locals {
  path = coalesce(var.path, "/")
  sts_roles = {
    role_arn         = local.installer_role_arn,
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Worker-Role"
    },
    operator_role_prefix = var.operator_role_prefix,
    oidc_config_id       = module.oidc_config.id
  }
}

data "aws_caller_identity" "current" {
}

resource "rhcs_cluster_rosa_classic" "rosa_sts_cluster" {
  name                = var.cluster_name
  cloud_region        = var.cloud_region
  aws_account_id      = data.aws_caller_identity.current.account_id
  availability_zones  = var.availability_zones
  replicas            = var.replicas
  autoscaling_enabled = var.autoscaling_enabled
  min_replicas        = var.min_replicas
  max_replicas        = var.max_replicas
  version             = var.openshift_version
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts = local.sts_roles
  wait_for_create_complete = true
}
