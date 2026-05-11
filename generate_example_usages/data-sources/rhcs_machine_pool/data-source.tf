# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

data "rhcs_machine_pool" "machine_pool" {
  cluster = "cluster-id-123"
  id = "my-pool"
}
