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

locals {
  path = coalesce(var.path, "/")
  sts_roles = {
    role_arn         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role${local.path}${var.account_role_prefix}-Worker-Role"
    },
    operator_role_prefix = var.operator_role_prefix,
    oidc_config_id       = rhcs_rosa_oidc_config.oidc_config.id
  }
}

data "aws_caller_identity" "current" {
}

# Wait 10 seconds
resource "time_sleep" "wait_10_seconds" {
  create_duration = "10s"
  depends_on = [ module.operator_roles_and_oidc_provider ]
}

# Generate random password for the cluster-admin
resource "random_password" "password" {
  length           = 23
  special          = false
  min_lower        = 1
  min_upper        = 1
  min_numeric      = 1
}

# Create HCP cluster
resource "rhcs_cluster_rosa_hcp" "rosa_hcp_cluster" {
  name                = var.cluster_name
  cloud_region        = var.cloud_region
  aws_account_id      = data.aws_caller_identity.current.account_id
  availability_zones  = var.availability_zones
  replicas            = var.replicas
  aws_subnet_ids      = var.aws_subnet_ids
  version             = var.openshift_version
  multi_az            = true
  admin_credentials   = {"username": var.username, "password": random_password.password.result}
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts = local.sts_roles
  depends_on = [ time_sleep.wait_10_seconds ]
}

# Wait for cluster current_compute == desired_compute
resource "rhcs_cluster_hcp_wait" "rosa_cluster_hcp" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  # timeout in minutes
  timeout = 60
}

# Get cluster URLs
data "rhcs_cluster_data" "cluster" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id
  depends_on = [ rhcs_cluster_hcp_wait.rosa_cluster_hcp ]
}

