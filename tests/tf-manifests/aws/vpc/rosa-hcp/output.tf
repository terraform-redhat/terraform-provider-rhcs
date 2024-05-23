output "cluster-private-subnet" {
  value = aws_subnet.private_subnet[*].id
}

output "cluster-public-subnet" {
  value = aws_subnet.public_subnet[*].id
}

output "node-private-subnet" {
  value = aws_subnet.private_subnet[*].id
}

output "azs" {
  value       = aws_subnet.private_subnet[*].availability_zone
}

output "vpc-id" {
  value       = time_sleep.vpc_resources_wait.triggers["vpc_id"]
}

output "vpc-cidr" {
  value = var.vpc_cidr
}
