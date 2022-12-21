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
      version = ">= 4.20.0"
    }
    ocm = {
      version = ">= 0.1"
      source  = "openshift-online/ocm"
    }
  }
}


provider "ocm" {
  token = var.token
  url = var.url
}

data "ocm_rosa_operator_roles" "operator_roles" {
  cluster_id =var.cluster_id
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix = var.account_role_prefix
}

module operator_roles {
    source  = "git::https://github.com/openshift-online/terraform-provider-ocm.git//modules/aws_roles"

    cluster_id = var.cluster_id
    rh_oidc_provider_thumbprint = var.oidc_thumbprint
    rh_oidc_provider_url = var.oidc_endpoint_url
    operator_roles_properties = data.ocm_rosa_operator_roles.operator_roles.operator_iam_roles
}
