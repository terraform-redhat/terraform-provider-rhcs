#
# Copyright (c) 2026 Red Hat, Inc.
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
      version = ">= 1.7.3"
      source  = "terraform-redhat/rhcs"
    }
  }
}

provider "rhcs" {
  url   = var.url
  token = var.token
}

resource "rhcs_hcp_machine_pool" "machine_pool" {
  cluster             = var.cluster_id
  name                = var.name
  replicas            = var.replicas
  labels              = var.labels
  auto_repair         = true
  subnet_id           = var.subnet_id
  aws_node_pool       = {
    instance_type     = var.machine_type
    image_type        = "Windows"
  }
  autoscaling = {
    enabled = var.autoscaling_enabled
  }
}
