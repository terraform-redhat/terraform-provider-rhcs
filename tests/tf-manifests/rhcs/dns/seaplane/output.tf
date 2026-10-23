# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "dns_domain_id" {
  value = rhcs_dns_domain_seaplane.dns_domain.id
}

output "dns_domain_cluster_arch" {
  value = rhcs_dns_domain_seaplane.dns_domain.cluster_arch
}
