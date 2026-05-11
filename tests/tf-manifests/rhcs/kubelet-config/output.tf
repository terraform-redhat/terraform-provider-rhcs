# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "kubelet_configs" {
  value = rhcs_kubeletconfig.kubeletconfig
}
