#
# Copyright (c) 2022 Red Hat, Inc.
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
      version = ">= 4.67.0"
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

module "create_account_roles" {
  source  = "terraform-redhat/rosa-classic/rhcs//modules/account-iam-resources"
  version = ">= 1.7.3"

  account_role_prefix = var.account_role_prefix
  openshift_version   = var.openshift_version
  path                = var.path
  tags                = var.tags
}
