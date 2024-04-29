terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform.local/local/rhcs"
    }
  }
}



resource "rhcs_cluster_rosa_classic" "rosa_sts_cluster_import" {
  aws_account_id = ""
  cloud_region   = ""
  name           = ""
}
resource "rhcs_cluster_rosa_classic" "rosa_import_no_cluster_id" {
  aws_account_id = ""
  cloud_region   = ""
  name           = ""
}
resource "rhcs_identity_provider" "idp_google_import" {
  name    = ""
  cluster = ""
  openid = {
    client_id     = ""
    client_secret = ""
    issuer        = ""
    claims = {
      email              = []
      groups             = []
      name               = []
      preferred_username = []
    }
  }
}

resource "rhcs_identity_provider" "idp_gitlab_import" {
  name    = ""
  cluster = ""
  openid = {
    client_id     = ""
    client_secret = ""
    issuer        = ""
    claims = {
      email              = []
      groups             = []
      name               = []
      preferred_username = []
    }
  }
}

resource "rhcs_machine_pool" "mp_import" {
  name         = ""
  machine_type = ""
  cluster      = ""
}

resource "rhcs_default_ingress" "default_ingress"{
  cluster = ""
  load_balancer_type = ""
  id = ""
}