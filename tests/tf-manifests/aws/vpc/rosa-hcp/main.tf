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
  shared_credentials_files = var.aws_shared_credentials_files
}

locals {
  availability_zones_count = var.multi_az ? 3 : 1
  tags = var.tags == null ? {} : var.tags

  private_subnet_tags = var.disable_subnet_tagging ? {} : { "kubernetes.io/role/internal-elb" = "" }
  public_subnet_tags  = var.disable_subnet_tagging ? {} : { "kubernetes.io/role/elb" = "" }
}

resource "aws_vpc" "vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = merge(
    {
      "Name" = var.name
    },
    local.tags
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.vpc.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

resource "aws_subnet" "public_subnet" {
  count = local.availability_zones_count

  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(var.vpc_cidr, local.availability_zones_count * 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]
  tags = merge(
    {
      "Name" = join("-", [var.name, "subnet", "public${count.index + 1}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.public_subnet_tags,
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_subnet" "private_subnet" {
  count = local.availability_zones_count

  vpc_id            = aws_vpc.vpc.id
  cidr_block        = cidrsubnet(var.vpc_cidr, local.availability_zones_count * 2, count.index + local.availability_zones_count)
  availability_zone = data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]
  tags = merge(
    {
      "Name" = join("-", [var.name, "subnet", "private${count.index + 1}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.private_subnet_tags,
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Internet gateway
#
resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = "${var.name}-igw"
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Elastic IPs for NAT gateways
#
resource "aws_eip" "eip" {
  count = local.availability_zones_count

  domain = "vpc"
  tags = merge(
    {
      "Name" = join("-", [var.name, "eip", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# NAT gateways
#
resource "aws_nat_gateway" "public_nat_gateway" {
  count = local.availability_zones_count

  allocation_id = aws_eip.eip[count.index].id
  subnet_id     = aws_subnet.public_subnet[count.index].id

  tags = merge(
    {
      "Name" = join("-", [var.name, "nat", "public${count.index}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Route tables
#
resource "aws_route_table" "public_route_table" {
  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = "${var.name}-public"
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_route_table" "private_route_table" {
  count = local.availability_zones_count

  vpc_id = aws_vpc.vpc.id
  tags = merge(
    {
      "Name" = join("-", [var.name, "rtb", "private${count.index}", data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]])
    },
    local.tags,
  )
  lifecycle {
    ignore_changes = [tags]
  }
}

#
# Routes
#
# Send all IPv4 traffic to the internet gateway
resource "aws_route" "ipv4_egress_route" {
  route_table_id         = aws_route_table.public_route_table.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
  depends_on             = [aws_route_table.public_route_table]
}

# Send all IPv6 traffic to the internet gateway
resource "aws_route" "ipv6_egress_route" {
  route_table_id              = aws_route_table.public_route_table.id
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.internet_gateway.id
  depends_on                  = [aws_route_table.public_route_table]
}

# Send private traffic to NAT
resource "aws_route" "private_nat" {
  count = local.availability_zones_count

  route_table_id         = aws_route_table.private_route_table[count.index].id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.public_nat_gateway[count.index].id
  depends_on             = [aws_route_table.private_route_table, aws_nat_gateway.public_nat_gateway]
}


# Private route for vpc endpoint
resource "aws_vpc_endpoint_route_table_association" "private_vpc_endpoint_route_table_association" {
  count = local.availability_zones_count

  route_table_id  = aws_route_table.private_route_table[count.index].id
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

#
# Route table associations
#
resource "aws_route_table_association" "public_route_table_association" {
  count = local.availability_zones_count

  subnet_id      = aws_subnet.public_subnet[count.index].id
  route_table_id = aws_route_table.public_route_table.id
}

resource "aws_route_table_association" "private_route_table_association" {
  count = local.availability_zones_count

  subnet_id      = aws_subnet.private_subnet[count.index].id
  route_table_id = aws_route_table.private_route_table[count.index].id
}

# This resource is used in order to add dependencies on all resources 
# Any resource uses this VPC ID, must wait to all resources creation completion
resource "time_sleep" "vpc_resources_wait" {
  create_duration = "20s"
  destroy_duration = "20s"
  triggers = {
    vpc_id                                           = aws_vpc.vpc.id
    cidr_block                                       = aws_vpc.vpc.cidr_block
    ipv4_egress_route_id                             = aws_route.ipv4_egress_route.id
    ipv6_egress_route_id                             = aws_route.ipv6_egress_route.id
    private_nat_ids                                  = jsonencode([for value in aws_route.private_nat : value.id])
    private_vpc_endpoint_route_table_association_ids = jsonencode([for value in aws_vpc_endpoint_route_table_association.private_vpc_endpoint_route_table_association : value.id])
    public_route_table_association_ids               = jsonencode([for value in aws_route_table_association.public_route_table_association : value.id])
    private_route_table_association_ids              = jsonencode([for value in aws_route_table_association.private_route_table_association : value.id])
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  # New configuration to exclude Local Zones
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}