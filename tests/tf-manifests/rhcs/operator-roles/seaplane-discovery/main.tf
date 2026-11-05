# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

# Discovery data source used by the import acceptance test.
data "rhcs_seaplane_operator_roles" "discovered" {
  operator_roles_prefix = var.operator_roles_prefix
}
