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

resource "rhcs_ocm_role_link" "ocm_role_link" {
  role_arn = var.role_arn
}
