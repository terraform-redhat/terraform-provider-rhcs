output "shared_role" {
  description = "Shared VPC Role ARN"
  value       = module.shared_vpc_policy_and_hosted_zone.shared_role
}

output "hosted_zone_id" {
  description = "Hosted Zone ID"
  value       = module.shared_vpc_policy_and_hosted_zone.hosted_zone_id
}

output "shared_subnets" {
  description = "The Amazon Resource Names (ARN) of the resource share"
  value       = module.shared_vpc_policy_and_hosted_zone.shared_subnets
}
