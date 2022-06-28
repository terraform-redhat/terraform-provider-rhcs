variable operator_role_prefix {
  type = string  
}

variable cluster_name {
  type = string
}

variable proxy {
  type = string
}

variable token {
  type = string
  sensitive = true
}

variable transit_gateway_id {
  type = string
}

terraform {
  required_providers {
    ocm = {
      source = "rh-mobb/ocm"
      version = "0.1.8"
    }
  }
}

provider "ocm" {
  token = var.token
}

data "aws_caller_identity" "current" {}

module openshift_vpc {
    source = "rh-mobb/rosa-privatelink-vpc/aws"
    name = "shading-private"
    region = "us-east-2"
    azs  = ["us-east-2a", "us-east-2b", "us-east-2c"]
    cidr = "10.0.0.0/22"
    private_subnets_cidrs = ["10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"]
    transit_gateway = {
        peer = true
        transit_gateway_id = var.transit_gateway_id
        dest_cidrs = ["192.168.246.0/23", "192.168.0.0/24"]
    }
}

locals {
  sts_roles = {
      role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Installer-Role",
      support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Support-Role",
      operator_iam_roles = [
        {
          name =  "cloud-credential-operator-iam-ro-creds",
          namespace = "openshift-cloud-credential-operator",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.operator_role_prefix}-openshift-cloud-credential-operator-cloud-c",
        },
        {
          name =  "installer-cloud-credentials",
          namespace = "openshift-image-registry",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.operator_role_prefix}-openshift-image-registry-installer-cloud-cr",
        },
        {
          name =  "cloud-credentials",
          namespace = "openshift-ingress-operator",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.operator_role_prefix}-openshift-ingress-operator-cloud-credential",
        },
        {
          name =  "ebs-cloud-credentials",
          namespace = "openshift-cluster-csi-drivers",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.operator_role_prefix}-openshift-cluster-csi-drivers-ebs-cloud-cre",
        },
        {
          name =  "cloud-credentials",
          namespace = "openshift-cloud-network-config-controller",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.operator_role_prefix}-openshift-cloud-network-config-controller-c",
        },
        {
          name =  "aws-cloud-credentials",
          namespace = "openshift-machine-api",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.operator_role_prefix}-openshift-machine-api-aws-cloud-credentials",
        },
      ]
      instance_iam_roles = {
        master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-ControlPlane-Role",
        worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Worker-Role"
      },    
  }
}

resource "ocm_cluster" "rosa_cluster" {
  name           = var.cluster_name
  cloud_provider = "aws"
  cloud_region   = "us-east-2"
  product        = "rosa"
  aws_account_id     = data.aws_caller_identity.current.account_id
  aws_subnet_ids = module.openshift_vpc.rosa_subnet_ids
  machine_cidr = module.openshift_vpc.rosa_vpc_cidr
  multi_az = true
  aws_private_link = true
  availability_zones = ["us-east-2a", "us-east-2b", "us-east-2c"]
  proxy = {
    http_proxy = var.proxy
    https_proxy = var.proxy
  }
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  wait = false
  sts = local.sts_roles
}

module sts_roles {
    source  = "rh-mobb/rosa-sts-roles/aws"
    create_account_roles = false
    clusters = [{
        id = ocm_cluster.rosa_cluster.id
        operator_role_prefix = var.operator_role_prefix
    }]
}
