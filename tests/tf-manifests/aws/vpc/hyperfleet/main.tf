# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# ── VPC ──────────────────────────────────────────────────────────────────────

resource "aws_vpc" "hyperfleet" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    Name    = "${var.name_prefix}-vpc"
    Cluster = var.name_prefix
  }
}

# ── Subnets ───────────────────────────────────────────────────────────────────

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.hyperfleet.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 3, 0)
  availability_zone = var.availability_zone
  tags = {
    Name                              = "${var.name_prefix}-private-subnet"
    "kubernetes.io/role/internal-elb" = "1"
    Cluster                           = var.name_prefix
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.hyperfleet.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, 8)
  availability_zone       = var.availability_zone
  map_public_ip_on_launch = true
  tags = {
    Name                     = "${var.name_prefix}-public-subnet"
    "kubernetes.io/role/elb" = "1"
    Cluster                  = var.name_prefix
  }
}

# ── Internet Gateway + NAT ────────────────────────────────────────────────────

resource "aws_internet_gateway" "hyperfleet" {
  vpc_id = aws_vpc.hyperfleet.id
  tags   = { Name = "${var.name_prefix}-igw", Cluster = var.name_prefix }
}

resource "aws_eip" "nat" {
  domain = "vpc"
  tags   = { Name = "${var.name_prefix}-nat-eip", Cluster = var.name_prefix }
}

resource "aws_nat_gateway" "hyperfleet" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public.id
  depends_on    = [aws_internet_gateway.hyperfleet]
  tags          = { Name = "${var.name_prefix}-natgw", Cluster = var.name_prefix }
}

# ── Route tables ──────────────────────────────────────────────────────────────

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.hyperfleet.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.hyperfleet.id
  }
  tags = { Name = "${var.name_prefix}-public-rtb", Cluster = var.name_prefix }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.hyperfleet.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.hyperfleet.id
  }
  tags = { Name = "${var.name_prefix}-private-rtb", Cluster = var.name_prefix }
}

resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}

# ── Route53 private hosted zone ───────────────────────────────────────────────

resource "aws_route53_zone" "hyperfleet" {
  name    = "${var.name_prefix}.hypershift.local"
  comment = "Private hosted zone for hyperfleet cluster ${var.name_prefix}"

  vpc {
    vpc_id = aws_vpc.hyperfleet.id
  }

  tags = {
    Name                                       = "${var.name_prefix}.hypershift.local"
    "kubernetes.io/cluster/${var.name_prefix}" = "owned"
    Cluster                                    = var.name_prefix
  }
}
