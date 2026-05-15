# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

data "rhcs_rosa_operator_roles" "operator_roles" {
  operator_role_prefix = "<operator-role-prefix>"
  account_role_prefix  = "<account-role-prefix>"
}