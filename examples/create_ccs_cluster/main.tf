#
# Copyright (c***REMOVED*** 2021 Red Hat, Inc.
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
      version = ">= 3.67"
    }
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform-redhat/rhcs"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

provider "rhcs" {
}

data "aws_caller_identity" "current" {
}

resource "aws_iam_user" "admin" {
  name = "osdCcsAdmin"
}

resource "aws_iam_user_policy_attachment" "admin_access" {
  user       = aws_iam_user.admin.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

resource "aws_iam_access_key" "admin_key" {
  user = aws_iam_user.admin.name
}

resource "rhcs_cluster" "my_cluster" {
  name                  = "my-cluster"
  cloud_provider        = "aws"
  cloud_region          = "us-east-1"
  ccs_enabled           = true
  aws_account_id        = data.aws_caller_identity.current.account_id
  aws_access_key_id     = aws_iam_access_key.admin_key.id
  aws_secret_access_key = aws_iam_access_key.admin_key.secret
}
