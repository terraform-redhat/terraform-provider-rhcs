terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform-redhat/rhcs"
    }
  }
}

provider "rhcs" {
}

# Create multiple image mirrors using for_each
resource "rhcs_image_mirror" "mirrors" {
  for_each = var.registry_mirrors

  cluster_id = var.cluster_id
  source     = each.key
  mirrors    = each.value
  type       = "digest"
}

# Data source to retrieve all image mirrors for verification
data "rhcs_image_mirrors" "cluster_mirrors" {
  cluster_id = var.cluster_id

  depends_on = [rhcs_image_mirror.mirrors]
}