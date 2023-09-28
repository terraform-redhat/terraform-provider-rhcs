#
# Copyright (c) 2023 Red Hat, Inc.
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
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}


provider "rhcs" {
  token = var.token
  url   = var.gateway
}

resource "rhcs_identity_provider" "htpasswd_idp" {
  cluster = var.cluster_id
  name    = var.name
  mapping_method = var.mapping_method
  htpasswd = {
    users = var.htpasswd_users
  }
}