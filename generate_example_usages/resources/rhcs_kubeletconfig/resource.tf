# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

# Example KubeletConfig
resource rhcs_kubeletconfig "example_kubeletconfig" {
  cluster = "cluster-id-123"
  pod_pids_limit = 10000
}