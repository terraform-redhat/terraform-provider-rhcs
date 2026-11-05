# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "role_count" {
  value = length(data.rhcs_seaplane_operator_roles.discovered.roles)
}

output "definition_version" {
  value = data.rhcs_seaplane_operator_roles.discovered.definition_version
}
