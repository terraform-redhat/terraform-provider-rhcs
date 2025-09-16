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

resource "rhcs_image_mirror" "image_mirror" {
  cluster_id = var.cluster
  type       = var.type
  source     = var.source_registry
  mirrors    = var.mirrors
}