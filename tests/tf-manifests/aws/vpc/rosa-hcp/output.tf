output "cluster_private_subnet" {
  value = aws_subnet.private_subnet[*].id
}

output "cluster_public_subnet" {
  value = aws_subnet.public_subnet[*].id
}

output "azs" {
  value = aws_subnet.private_subnet[*].availability_zone
}

<<<<<<< HEAD
output "vpc_id" {
  value       = time_sleep.vpc_resources_wait.triggers["vpc_id"]
=======
output "vpc-id" {
  value = time_sleep.vpc_resources_wait.triggers["vpc_id"]
>>>>>>> f199045 (OCM-8537 | ci: Automated 74520 and updated 73431,72514,70129,70128)
}

output "vpc_cidr" {
  value = var.vpc_cidr
}
