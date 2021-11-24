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
    ocm = {
      version = ">= 0.1"
      source  = "openshift-online/ocm"
    }
  }
}

provider "ocm" {
}

resource "ocm_cluster" "my_cluster" {
  name           = "my-cluster"
  cloud_provider = "aws"
  cloud_region   = "us-east-1"
  compute_nodes  = 10
  version        = "openshift-v4.9.7"
  properties = {
    department  = "accounting"
    application = "billing"
  }
}

resource "ocm_identity_provider" "my_idp" {
  cluster = ocm_cluster.my_cluster.id
  name    = "my-idp"
  htpasswd = {
    username = "admin"
    password = "redhat123"
  }
}

resource "ocm_group_membership" "my_admin" {
  cluster = ocm_cluster.my_cluster.id
  group   = "dedicated-admins"
  user    = "admin"
}

resource "ocm_machine_pool" "my_pool" {
  cluster      = ocm_cluster.my_cluster.id
  name         = "my-pool"
  machine_type = "r5.xlarge"
  replicas     = 5
}
