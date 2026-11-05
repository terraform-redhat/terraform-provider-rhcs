# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "definition_version" {
  value = rhcs_seaplane_operator_roles.roles.definition_version
}

output "instance_profile_name" {
  value = rhcs_seaplane_operator_roles.roles.instance_profile_name
}

output "worker_role_arn" {
  value = [for role in rhcs_seaplane_operator_roles.roles.roles : role.arn if role.is_worker_role][0]
}

output "role_count" {
  value = length(rhcs_seaplane_operator_roles.roles.roles)
}
