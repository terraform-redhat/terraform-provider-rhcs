#
# Copyright (c***REMOVED*** 2023 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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
    ocm = {
      version = ">=1.0.2"
      source  = "terraform-redhat/ocm"
    }
  }
}
provider "ocm" {
  token = var.token
  url = var.url
}

# Generates the OIDC config resources' names
resource "ocm_rosa_oidc_config_input" "oidc_input" {
  region = "us-east-2"
}

# Create the OIDC config resources on AWS
module oidc_config_input_resources {
  source = "terraform-redhat/rosa-sts/aws"
  version = "0.0.5"

  create_oidc_config_resources = true

  bucket_name = ocm_rosa_oidc_config_input.oidc_input.bucket_name
  discovery_doc = ocm_rosa_oidc_config_input.oidc_input.discovery_doc
  jwks = ocm_rosa_oidc_config_input.oidc_input.jwks
  private_key = ocm_rosa_oidc_config_input.oidc_input.private_key
  private_key_file_name = ocm_rosa_oidc_config_input.oidc_input.private_key_file_name
  private_key_secret_name = ocm_rosa_oidc_config_input.oidc_input.private_key_secret_name
}

# Create unmanaged OIDC config
resource "ocm_rosa_oidc_config" "oidc_config" {
  managed = false
  secret_arn =  module.oidc_config_input_resources.secret_arn
  issuer_url = ocm_rosa_oidc_config_input.oidc_input.issuer_url
  installer_role_arn = var.installer_role_arn
}

data "ocm_rosa_operator_roles" "operator_roles" {
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix = var.account_role_prefix
}

module operator_roles_and_oidc_provider {
  source = "terraform-redhat/rosa-sts/aws"
  version = "0.0.5"

  create_operator_roles = true
  create_oidc_provider = true

  cluster_id = ""
  rh_oidc_provider_thumbprint = ocm_rosa_oidc_config.oidc_config.thumbprint
  rh_oidc_provider_url = ocm_rosa_oidc_config.oidc_config.oidc_endpoint_url
  operator_roles_properties = data.ocm_rosa_operator_roles.operator_roles.operator_iam_roles
}

locals {
  sts_roles = {
    role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Worker-Role"
    },
    operator_role_prefix = var.operator_role_prefix,
    oidc_config_id = ocm_rosa_oidc_config.oidc_config.id
  }
}

data "aws_caller_identity" "current" {
}

resource "ocm_cluster_rosa_classic" "rosa_sts_cluster" {
  name           = "tf-gdb-test"
  cloud_region   = "us-east-2"
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = ["us-east-2a"]
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts = local.sts_roles
}

resource "ocm_cluster_wait" "rosa_cluster" {
  cluster = ocm_cluster_rosa_classic.rosa_sts_cluster.id
  # timeout in minutes
  timeout = 60
}