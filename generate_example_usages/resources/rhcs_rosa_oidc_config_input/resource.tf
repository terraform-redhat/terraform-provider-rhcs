# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

# Generates the OIDC config resources' names
resource "rhcs_rosa_oidc_config_input" "oidc_input" {
  region = "us-east-2"
}
