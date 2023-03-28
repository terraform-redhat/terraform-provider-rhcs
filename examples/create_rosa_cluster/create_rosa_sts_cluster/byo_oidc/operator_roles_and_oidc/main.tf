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
      version = "0.0.3"
      source  = "terraform-redhat/ocm"
    }
  }
}

provider "ocm" {
  token = var.token
  url = var.url
}

data "ocm_rosa_operator_roles" "rosa_sts" {
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix = var.account_role_prefix
}

module rosa_sts {
    source = "terraform-redhat/rosa-sts/aws"
    version = "0.0.4"

    cluster_id = var.cluster_id
    rh_oidc_provider_url = var.oidc_endpoint_url
    operator_roles_properties = data.ocm_rosa_operator_roles.rosa_sts.operator_iam_roles

    rh_oidc_provider_thumbprint = var.oidc_thumbprint
    create_oidc_provider = var.create_oidc_provider
}
