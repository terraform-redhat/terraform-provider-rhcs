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
      source  = "terraform-redhat/rhcs"
    }
  }
}

provider "rhcs" {
  token = var.token
  url   = var.url
}

resource "rhcs_machine_pool" "machine_pool" {
  cluster             = var.cluster_id
  name                = var.name
  machine_type        = var.machine_type
  replicas            = var.replicas
  autoscaling_enabled = var.autoscaling_enabled
  min_replicas        = var.min_replicas
  max_replicas        = var.max_replicas
  labels              = var.labels
}
