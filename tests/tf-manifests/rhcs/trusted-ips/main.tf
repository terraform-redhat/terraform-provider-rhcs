terraform {
  required_providers {
    rhcs = {
      version = ">= 1.0.1-0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

data "rhcs_trusted_ip_addresses" "all" {
}