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
data "rhcs_info" "info" {}