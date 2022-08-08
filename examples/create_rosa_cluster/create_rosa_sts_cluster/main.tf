terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    ocm = {
      version = ">= 0.1"
      source  = "openshift-online/ocm"
    }
  }
}

variable token {
  type = string
  sensitive = true
}

variable operator_role_prefix {
    type = string
}

provider "ocm" {
  token = var.token
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

data "aws_caller_identity" "current" {
}

resource "ocm_cluster" "rosa_cluster" {
  name           = "my-cluster"
  cloud_provider = "aws"
  cloud_region   = "us-east-2"
  product        = "rosa"
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = ["us-east-2a"]
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
