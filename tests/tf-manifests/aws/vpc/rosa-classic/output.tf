output "cluster_private_subnet" {
  value = module.vpc.private_subnets
}

output "cluster_public_subnet" {
  value = module.vpc.public_subnets
}

output "azs" {
  value = module.vpc.azs
}

output "vpc_id" {
  value = module.vpc.vpc_id
}

output "vpc_cidr" {
  value = var.vpc_cidr
}
