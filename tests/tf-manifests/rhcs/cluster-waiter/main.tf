terraform {
  required_providers {
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource "rhcs_cluster_wait" "rosa_cluster" {
  cluster = var.cluster_id
  timeout = var.timeout_in_min # in minutes
}
