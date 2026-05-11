# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

data "rhcs_hcp_machine_pool" "machine_pool" {
  cluster = "cluster-id-123"
  name = "my-pool"
}
