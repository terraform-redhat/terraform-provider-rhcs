terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
  }
}

provider "aws" {
  region                   = var.aws_region
  shared_credentials_files = var.aws_shared_credentials_files
}

# Standard VPC via module (creates public+private subnets with NAT Gateways)
module "vpc" {
  count   = var.no_nat_gateway ? 0 : 1
  source  = "terraform-redhat/rosa-classic/rhcs//modules/vpc"
  version = ">=1.6.3"

  vpc_cidr                 = var.vpc_cidr
  name_prefix              = var.name_prefix
  availability_zones       = var.availability_zones
  availability_zones_count = var.availability_zones_count
  tags                     = var.tags
}

# VPC with public+private subnets but NO NAT Gateways — used when no_nat_gateway = true.
# Uses an Internet Gateway (free, no quota) for public subnet routing.
# Satisfies OCM's 6-subnet requirement for Multi-AZ without consuming NAT Gateway quota.
data "aws_availability_zones" "available" {
  count = var.no_nat_gateway ? 1 : 0
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  no_nat_requested_count = var.availability_zones != null ? length(var.availability_zones) : (var.availability_zones_count != null ? var.availability_zones_count : 1)
  no_nat_az_count        = var.no_nat_gateway && var.availability_zones == null ? min(local.no_nat_requested_count, length(data.aws_availability_zones.available[0].names)) : local.no_nat_requested_count
  no_nat_azs             = var.no_nat_gateway ? (var.availability_zones != null ? var.availability_zones : slice(data.aws_availability_zones.available[0].names, 0, local.no_nat_az_count)) : []
}

resource "aws_vpc" "no_nat" {
  count                = var.no_nat_gateway ? 1 : 0
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-vpc"
  })
}

resource "aws_internet_gateway" "no_nat" {
  count  = var.no_nat_gateway ? 1 : 0
  vpc_id = aws_vpc.no_nat[0].id
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-igw"
  })
}

resource "aws_subnet" "no_nat_public" {
  count             = var.no_nat_gateway ? local.no_nat_az_count : 0
  vpc_id            = aws_vpc.no_nat[0].id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone = local.no_nat_azs[count.index]
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-public-${count.index}"
  })
}

resource "aws_subnet" "no_nat_private" {
  count             = var.no_nat_gateway ? local.no_nat_az_count : 0
  vpc_id            = aws_vpc.no_nat[0].id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + local.no_nat_az_count)
  availability_zone = local.no_nat_azs[count.index]
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-private-${count.index}"
  })
}

resource "aws_route_table" "no_nat_public" {
  count  = var.no_nat_gateway ? 1 : 0
  vpc_id = aws_vpc.no_nat[0].id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.no_nat[0].id
  }
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-public-rt"
  })
}

resource "aws_route_table" "no_nat_private" {
  count  = var.no_nat_gateway ? 1 : 0
  vpc_id = aws_vpc.no_nat[0].id
  tags = merge(var.tags, {
    Name = "${var.name_prefix}-private-rt"
  })
}

resource "aws_route_table_association" "no_nat_public" {
  count          = var.no_nat_gateway ? local.no_nat_az_count : 0
  subnet_id      = aws_subnet.no_nat_public[count.index].id
  route_table_id = aws_route_table.no_nat_public[0].id
}

resource "aws_route_table_association" "no_nat_private" {
  count          = var.no_nat_gateway ? local.no_nat_az_count : 0
  subnet_id      = aws_subnet.no_nat_private[count.index].id
  route_table_id = aws_route_table.no_nat_private[0].id
}

locals {
  out_vpc_id             = var.no_nat_gateway ? aws_vpc.no_nat[0].id : module.vpc[0].vpc_id
  out_vpc_cidr           = var.no_nat_gateway ? aws_vpc.no_nat[0].cidr_block : module.vpc[0].cidr_block
  out_private_subnets    = var.no_nat_gateway ? [for s in aws_subnet.no_nat_private : s.id] : module.vpc[0].private_subnets
  out_public_subnets     = var.no_nat_gateway ? [for s in aws_subnet.no_nat_public : s.id] : module.vpc[0].public_subnets
  out_availability_zones = var.no_nat_gateway ? [for s in aws_subnet.no_nat_private : s.availability_zone] : module.vpc[0].availability_zones
}
