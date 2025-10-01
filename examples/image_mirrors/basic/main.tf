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

# Create a simple image mirror for nginx
resource "rhcs_image_mirror" "nginx_mirror" {
  cluster_id = var.cluster_id
  source     = var.source_registry
  mirrors    = var.mirrors
  type       = var.type
}

# Data source to retrieve all image mirrors for the cluster
data "rhcs_image_mirrors" "cluster_mirrors" {
  cluster_id = var.cluster_id

  depends_on = [rhcs_image_mirror.nginx_mirror]
}