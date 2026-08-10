# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "oidc_provider_arn" {
  value = aws_iam_openid_connect_provider.hyperfleet.arn
}

output "worker_role_arn" {
  value = aws_iam_role.worker.arn
}

output "worker_instance_profile_name" {
  value = aws_iam_instance_profile.worker.name
}
