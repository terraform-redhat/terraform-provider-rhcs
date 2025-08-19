terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0-0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

data "rhcs_info" "info" {}