# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "trusted_ips" {
  value = data.rhcs_trusted_ip_addresses.all
}
