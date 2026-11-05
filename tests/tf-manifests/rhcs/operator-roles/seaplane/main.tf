# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0
#
# Creates the full set of Seaplane V2 operator roles (including the worker
# instance role) via the rhcs_seaplane_operator_roles resource. Apply after
# an OIDC config exists so oidc_issuer_url is known.

terraform {
  required_providers {
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

data "aws_caller_identity" "current" {}

provider "rhcs" {
  seaplane_url   = var.seaplane_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
}

resource "rhcs_seaplane_operator_roles" "roles" {
  operator_roles_prefix = var.operator_roles_prefix
  oidc_issuer_url       = var.oidc_issuer_url
}
