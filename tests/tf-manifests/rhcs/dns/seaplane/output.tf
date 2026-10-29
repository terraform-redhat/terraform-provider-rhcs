# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "dns_domain_id" {
  value = rhcs_dns_domain_seaplane.dns_domain.id
}

output "base_domain" {
  value = rhcs_dns_domain_seaplane.dns_domain.base_domain
}

output "phase" {
  value = rhcs_dns_domain_seaplane.dns_domain.phase
}
