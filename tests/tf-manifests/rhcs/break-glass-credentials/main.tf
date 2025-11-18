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

resource "rhcs_break_glass_credential" "rosa_break_glass_credential" {
  cluster             = var.cluster
  expiration_duration = var.expiration_duration
  username            = var.username
}
