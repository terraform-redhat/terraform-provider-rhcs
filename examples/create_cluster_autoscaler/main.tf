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

resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
  cluster                       = var.cluster_id
  balance_similar_node_groups   = true
  skip_nodes_with_local_storage = true
  log_verbosity                 = 1
  max_pod_grace_period          = 10
  pod_priority_threshold        = -10
  ignore_daemonsets_utilization = true
  max_nodes_provision_time      = "1h"
  balancing_ignored_labels      = ["l1", "l2"]

  resource_limits = {
    max_nodes_total = 2
    cores = {
      min = 0
      max = 1
    }
    memory = {
      min = 0
      max = 1
    }
    gpus = [
      {
        type = "nvidia"
        range = {
          min = 0
          max = 1
        }
      },
      {
        type = "intel"
        range = {
          min = 1
          max = 2
        }
      },
    ]
  }

  scale_down = {
    enabled               = true
    utilization_threshold = "0.4"
    unneeded_time         = "1h"
    delay_after_add       = "3h"
    delay_after_delete    = "3h"
    delay_after_failure   = "3h"
  }
}
