output "shared_role" {
  description = "Shared VPC Role ARN"
  value       = time_sleep.shared_resources_propagation.triggers["shared_vpc_role_arn"]
}

output "hosted_zone_id" {
  description = "Hosted Zone ID"
  value       = time_sleep.shared_resources_propagation.triggers["shared_vpc_hosted_zone_id"]
}

output "shared_subnets" {
  description = "The Amazon Resource Names (ARN) of the resource share"
  value       = [for resource_arn in aws_ram_resource_association.shared_vpc_resource_association[*].resource_arn : trimprefix(resource_arn, local.resource_arn_prefix)]
}

output "azs" {
  value = distinct(values(data.aws_subnet.shared_subnets)[*].availability_zone)
}