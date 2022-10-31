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

variable token {
  type = string
  sensitive = true
}

variable account_role_prefix {
    type = string
}

variable operator_role_prefix {
    type = string
}

provider "ocm" {
  token = var.token
}

data "aws_caller_identity" "current" {
}

locals {
  sts_vars = {
    operator_role_prefix = "terraform-ocm"
    account_role_prefix = "ManagedOpenshift"
  }
}

resource "ocm_cluster" "rosa_cluster" {
  name           = "my-cluster"
  cloud_provider = "aws"
  cloud_region   = "us-east-2"
  product        = "rosa"
  availability_zones = ["us-east-2a"]
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  wait = false
}

module sts_roles {
    source  = "rh-mobb/rosa-sts-roles/aws"
    create_account_roles = false
    clusters = [{
        id = ocm_cluster.rosa_cluster.id
        operator_role_prefix = var.operator_role_prefix
    }]
    rh_oidc_provider_thumbprint = ocm_cluster.rosa_cluster.thumbprint
    rh_oidc_provider_url = ocm_cluster.rosa_cluster.sts.oidc_endpoint_url
}
