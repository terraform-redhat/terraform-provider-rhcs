output "cluster-private-subnet" {
  value = module.vpc.private_subnets
}

output "cluster-public-subnet" {
  value = module.vpc.public_subnets
}

output "node-private-subnet" {
  value = module.vpc.private_subnets
}

output "azs"{
  value = module.vpc.azs
}

output "vpc-cidr"{
  value = var.vpc_cidr
}
