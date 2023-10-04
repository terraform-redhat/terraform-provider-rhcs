terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  token = var.token
  url   = var.url
}

resource "rhcs_machine_pool" "mp" {
  cluster      = var.cluster
  machine_type = var.machine_type
  name         = var.name
  replicas     = var.replicas
  labels       = var.labels
  taints       = var.taints
}
