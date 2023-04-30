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

resource "ocm_rosa_oidc_config" "oidc_config" {
  managed = true
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
