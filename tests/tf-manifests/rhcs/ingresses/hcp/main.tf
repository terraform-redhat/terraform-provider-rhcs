terraform {
  required_providers {
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  url = var.url
}

resource "rhcs_hcp_default_ingress" "default_ingress" {
  cluster          = var.cluster
  listening_method = var.listening_method
}
