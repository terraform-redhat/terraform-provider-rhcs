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
      version = ">=1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}
provider "rhcs" {
  token = var.token
  url   = var.url
}

# Generates the OIDC config resources' names
resource "rhcs_rosa_oidc_config_input" "oidc_input" {
  region = var.aws_region
}

provider "aws" {
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  region     = var.aws_region
}

# Create the OIDC config resources on AWS
module "oidc_config_input_resources" {
  source  = "terraform-redhat/rosa-sts/aws"
  version = "0.0.14"

  create_oidc_config_resources = true

  bucket_name             = rhcs_rosa_oidc_config_input.oidc_input.bucket_name
  discovery_doc           = rhcs_rosa_oidc_config_input.oidc_input.discovery_doc
  jwks                    = rhcs_rosa_oidc_config_input.oidc_input.jwks
  private_key             = rhcs_rosa_oidc_config_input.oidc_input.private_key
  private_key_file_name   = rhcs_rosa_oidc_config_input.oidc_input.private_key_file_name
  private_key_secret_name = rhcs_rosa_oidc_config_input.oidc_input.private_key_secret_name
}

# Create unmanaged OIDC config
resource "rhcs_rosa_oidc_config" "oidc_config" {
  managed            = false
  secret_arn         = module.oidc_config_input_resources.secret_arn
  issuer_url         = rhcs_rosa_oidc_config_input.oidc_input.issuer_url
  installer_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Installer-Role"
}

data "rhcs_rosa_operator_roles" "operator_roles" {
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix  = var.account_role_prefix
}

module "operator_roles_and_oidc_provider" {
  source  = "terraform-redhat/rosa-sts/aws"
  version = "0.0.14"

  create_operator_roles = true
  create_oidc_provider  = true

  cluster_id                  = ""
  rh_oidc_provider_thumbprint = rhcs_rosa_oidc_config.oidc_config.thumbprint
  rh_oidc_provider_url        = rhcs_rosa_oidc_config.oidc_config.oidc_endpoint_url
  operator_roles_properties   = data.rhcs_rosa_operator_roles.operator_roles.operator_iam_roles
}

locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Worker-Role"
    },
    operator_role_prefix = var.operator_role_prefix,
    oidc_config_id       = rhcs_rosa_oidc_config.oidc_config.id
  }
}

data "aws_caller_identity" "current" {
}

resource "rhcs_cluster_rosa_classic" "rosa_sts_cluster" {
  name               = var.cluster_name
  version            = var.openshift_version
  channel_group      = var.channel_group
  cloud_region       = var.aws_region
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = var.aws_availability_zones
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts             = local.sts_roles
  replicas        = var.replicas
  destroy_timeout = 120
  wait_for_create_complete = true
}
