terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
}

resource rhcs_break_glass_credential rosa_break_glass_credential {
  cluster             = var.cluster
  expiration_duration = var.expiration_duration
  username            = var.username
}
