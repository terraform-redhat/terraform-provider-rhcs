terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  url   = var.url
}

resource "rhcs_machine_pool" "mp" {
 cluster                 = var.cluster
  machine_type            = var.machine_type
  name                    = var.name
  replicas                = var.replicas
  labels                  = var.labels
  taints                  = var.taints
  min_replicas            = var.min_replicas
  max_replicas            = var.max_replicas
  autoscaling_enabled     = var.autoscaling_enabled
  availability_zone       = var.availability_zone
  subnet_id               = var.subnet_id
  multi_availability_zone = var.multi_availability_zone

}
