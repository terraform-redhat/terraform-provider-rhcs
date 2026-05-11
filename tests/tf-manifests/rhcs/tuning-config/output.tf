# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "names" {
  value = rhcs_tuning_config.tcs[*].name
}

output "specs" {
  value = rhcs_tuning_config.tcs[*].spec
}