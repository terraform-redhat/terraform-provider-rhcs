# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "idp_id" {
  value = rhcs_identity_provider.htpasswd_idp.id
}