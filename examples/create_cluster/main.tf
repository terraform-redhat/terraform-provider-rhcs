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
      source  = "localhost/redhat/ocm"
    }
  }
}

provider "ocm" {
  url = "https://api.stage.openshift.com"
}

resource "ocm_cluster" "my_cluster" {
  name           = "my-cluster"
  cloud_provider = "aws"
  cloud_region   = "us-east-1"
  multi_az       = true
  properties = {
    fake_cluster = "true"
  }
}

resource "ocm_identity_provider" "my_htpasswd" {
  cluster_id = ocm_cluster.my_cluster.id
  name       = "my-htpasswd"
  htpasswd = {
    username = "my-user"
    password = "my-password"
  }
}

resource "ocm_identity_provider" "my_ldap" {
  cluster_id = ocm_cluster.my_cluster.id
  name       = "my-ldap"
  ldap = {
    bind_dn       = "my-bind-dn"
    bind_password = "my-bind-password"
    url           = "https://my-server.com"
    attributes = {
      id                 = ["my-id"]
      email              = ["my-email"]
      name               = ["my-name"]
      preferred_username = ["my-preferred-username"]
    }
  }
}
