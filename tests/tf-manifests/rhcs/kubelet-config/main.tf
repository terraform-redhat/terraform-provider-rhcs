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

resource "rhcs_kubeletconfig" "kubeletconfig" {
  cluster        = var.cluster
  pod_pids_limit = var.pod_pids_limit
}
