output "cluster_private_subnet" {
  value = aws_subnet.private_subnet[*].id
}

output "cluster_public_subnet" {
  value = aws_subnet.public_subnet[*].id
}

output "azs" {
  value = aws_subnet.private_subnet[*].availability_zone
}

output "vpc_id" {
  value = time_sleep.vpc_resources_wait.triggers["vpc_id"]
}

output "vpc_cidr" {
  value = var.vpc_cidr
}
