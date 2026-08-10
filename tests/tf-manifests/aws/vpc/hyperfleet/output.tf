# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

output "vpc_id" {
  value = aws_vpc.hyperfleet.id
}

output "private_subnet_id" {
  value = aws_subnet.private.id
}

output "public_subnet_id" {
  value = aws_subnet.public.id
}

output "availability_zone" {
  value = aws_subnet.private.availability_zone
}

output "hosted_zone_id" {
  value = aws_route53_zone.hyperfleet.id
}
