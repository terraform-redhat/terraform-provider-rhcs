# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    rhcs = {
      source  = "terraform.local/local/rhcs"
      version = ">= 1.1.0"
    }
  }
}

provider "rhcs" {}

resource "rhcs_external_auth_provider" "main" {
  cluster                      = var.cluster
  name                         = var.name
  issuer_url                   = var.issuer_url
  issuer_audiences             = var.issuer_audiences
  claim_mapping_groups_claim   = var.claim_mapping_groups_claim
  claim_mapping_username_claim = var.claim_mapping_username_claim
  claim_validation_rule        = var.claim_validation_rule
}
