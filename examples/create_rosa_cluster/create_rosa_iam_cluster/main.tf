#
# Copyright (c***REMOVED*** 2022 Red Hat, Inc.
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

provider "aws" {
  region = "us-east-1"
}

variable "token" {
  type      = string
  sensitive = true
}

provider "ocm" {
  token = var.token
}

data "aws_caller_identity" "current" {
}

data "aws_iam_user" "admin" {
  user_name = "osdCcsAdmin"
}

resource "aws_iam_access_key" "admin_key" {
  user = data.aws_iam_user.admin.user_name
}

resource "ocm_cluster" "rosa_cluster" {
  name                  = "my-cluster"
  cloud_provider        = "aws"
  cloud_region          = "us-east-1"
  product               = "rosa"
  aws_account_id        = data.aws_caller_identity.current.account_id
  availability_zones    = ["us-east-1a"]
  aws_access_key_id     = aws_iam_access_key.admin_key.id
  aws_secret_access_key = aws_iam_access_key.admin_key.secret
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
}
