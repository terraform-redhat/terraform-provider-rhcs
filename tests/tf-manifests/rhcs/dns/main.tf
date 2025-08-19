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

resource "rhcs_dns_domain" "dns_domain" {}
