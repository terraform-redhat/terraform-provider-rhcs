# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "sg_ids" {
  description = "created SG"
  value       = module.web_server_sg.*.security_group_id
}