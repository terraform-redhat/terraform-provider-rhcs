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
      source  = "terraform-redhat/ocm"
    }
  }
}

provider "ocm" {
  token = var.token
  url = var.url
}

locals {
  sts_roles = {
      role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Installer-Role",
      support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Support-Role",
      instance_iam_roles = {
        master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-ControlPlane-Role",
        worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Worker-Role"
      },
      operator_role_prefix = var.operator_role_prefix,
  }
}

data "aws_caller_identity" "current" {
}

resource "ocm_cluster_rosa_classic" "rosa_sts_cluster" {
  name           = "my-cluster"
  cloud_region   = "us-east-2"
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = ["us-east-2a"]
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts = local.sts_roles
}
