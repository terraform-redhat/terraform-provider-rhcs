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
      version = ">= 6.0.0"
    }
    rhcs = {
      version = ">= 1.6.2"
      source  = "terraform-redhat/rhcs"
    }
  }
}
provider "rhcs" {
  token = var.token
  url   = var.url
}

provider "aws" {
  region = var.cloud_region
}

module "oidc_config_and_provider" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/oidc-config-and-provider"
  version = ">= 1.7.3"

  managed            = var.managed
  installer_role_arn = var.installer_role_arn
  tags               = var.tags
}

module "operator_roles" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/operator-roles"
  version = ">= 1.7.3"

  operator_role_prefix = var.operator_role_prefix
  account_role_prefix  = var.account_role_prefix
  oidc_endpoint_url    = module.oidc_config_and_provider.oidc_endpoint_url
  tags                 = var.tags
  path                 = var.path
}
