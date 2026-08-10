# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0
#
# Creates the OIDC provider and all IAM roles required by a hyperfleet cluster.
# Apply AFTER the cluster resource so that oidc_issuer_url is known.

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = ">= 4.0.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

locals {
  partition  = data.aws_partition.current.partition
  account_id = data.aws_caller_identity.current.account_id

  # Strip the https:// prefix to get the provider identifier used in ARNs and
  # condition keys (e.g. s3.amazonaws.com/bucket/prefix).
  oidc_provider = trimprefix(var.oidc_issuer_url, "https://")

  operator_role_suffixes = [
    "-ingress",
    "-cloud-controller-manager",
    "-ebs-csi",
    "-image-registry",
    "-network-config",
    "-control-plane-operator",
    "-node-pool-management",
  ]

  operator_policies = {
    "-ingress"                  = "ROSAIngressOperatorPolicy"
    "-cloud-controller-manager" = "ROSAKubeControllerPolicy"
    "-ebs-csi"                  = "ROSAAmazonEBSCSIDriverOperatorPolicy"
    "-image-registry"           = "ROSAImageRegistryOperatorPolicy"
    "-network-config"           = "ROSACloudNetworkConfigOperatorPolicy"
    "-control-plane-operator"   = "ROSAControlPlaneOperatorPolicy"
    "-node-pool-management"     = "ROSANodePoolManagementPolicy"
  }

  # Service accounts allowed to assume each operator role (used in OIDC conditions).
  operator_service_accounts = {
    "-ingress" = [
      "system:serviceaccount:openshift-ingress-operator:ingress-operator",
    ]
    "-cloud-controller-manager" = [
      "system:serviceaccount:kube-system:kube-controller-manager",
    ]
    "-ebs-csi" = [
      "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
      "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
    ]
    "-image-registry" = [
      "system:serviceaccount:openshift-image-registry:cluster-image-registry-operator",
      "system:serviceaccount:openshift-image-registry:registry",
    ]
    "-network-config" = [
      "system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller",
    ]
    "-control-plane-operator" = [
      "system:serviceaccount:kube-system:control-plane-operator",
    ]
    "-node-pool-management" = [
      "system:serviceaccount:kube-system:capa-controller-manager",
    ]
  }

  worker_role_name = "${var.operator_roles_prefix}-ROSA-Worker-Role"
}

# ── OIDC provider ─────────────────────────────────────────────────────────────

data "tls_certificate" "oidc" {
  url = var.oidc_issuer_url
}

resource "aws_iam_openid_connect_provider" "hyperfleet" {
  url             = var.oidc_issuer_url
  client_id_list  = ["openshift"]
  thumbprint_list = [data.tls_certificate.oidc.certificates[0].sha1_fingerprint]
}

# ── Operator IAM roles ────────────────────────────────────────────────────────

data "aws_iam_policy_document" "operator_trust" {
  for_each = toset(local.operator_role_suffixes)

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.hyperfleet.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_provider}:sub"
      values   = local.operator_service_accounts[each.key]
    }
  }
}

resource "aws_iam_role" "operator" {
  for_each = toset(local.operator_role_suffixes)

  name               = "${var.operator_roles_prefix}${each.key}"
  assume_role_policy = data.aws_iam_policy_document.operator_trust[each.key].json
  description        = "ROSA HCP operator role for hyperfleet cluster ${var.operator_roles_prefix}"
  tags               = { Cluster = var.operator_roles_prefix }
}

resource "aws_iam_role_policy_attachment" "operator" {
  for_each = toset(local.operator_role_suffixes)

  role       = aws_iam_role.operator[each.key].name
  policy_arn = "arn:${local.partition}:iam::aws:policy/service-role/${local.operator_policies[each.key]}"
}

# ── Worker IAM role + instance profile ───────────────────────────────────────

data "aws_iam_policy_document" "worker_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "worker" {
  name               = local.worker_role_name
  assume_role_policy = data.aws_iam_policy_document.worker_trust.json
  description        = "ROSA HCP worker node role for hyperfleet cluster ${var.operator_roles_prefix}"
  tags               = { Cluster = var.operator_roles_prefix }
}

resource "aws_iam_role_policy_attachment" "worker_instance_policy" {
  role       = aws_iam_role.worker.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/service-role/ROSAWorkerInstancePolicy"
}

resource "aws_iam_role_policy_attachment" "worker_ssm_policy" {
  role       = aws_iam_role.worker.name
  policy_arn = "arn:${local.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "worker" {
  name = local.worker_role_name
  role = aws_iam_role.worker.name
}
