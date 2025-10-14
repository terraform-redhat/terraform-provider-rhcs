terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource "rhcs_external_auth_provider" "external_auth_provider" {
  cluster = var.cluster
  id      = var.id

  issuer = {
    url       = var.issuer_url
    audiences = var.issuer_audiences
    ca        = var.issuer_ca
  }

  dynamic "clients" {
    for_each = var.console_client_id != null && var.console_client_secret != null ? [1] : []
    content {
      id     = var.console_client_id
      secret = var.console_client_secret
    }
  }

  dynamic "claim" {
    for_each = (var.claim_mapping_username_key != null || var.claim_mapping_groups_key != null || length(var.claim_validation_rules) > 0) ? [1] : []
    content {
      dynamic "mappings" {
        for_each = (var.claim_mapping_username_key != null || var.claim_mapping_groups_key != null) ? [1] : []
        content {
          dynamic "username" {
            for_each = var.claim_mapping_username_key != null ? [1] : []
            content {
              claim = var.claim_mapping_username_key
            }
          }
          dynamic "groups" {
            for_each = var.claim_mapping_groups_key != null ? [1] : []
            content {
              claim = var.claim_mapping_groups_key
            }
          }
        }
      }
      
      dynamic "validation_rules" {
        for_each = var.claim_validation_rules
        content {
          claim          = validation_rules.value.claim
          required_value = validation_rules.value.required_value
        }
      }
    }
  }
}