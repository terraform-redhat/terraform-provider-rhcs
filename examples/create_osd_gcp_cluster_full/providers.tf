terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
    rhcs = {
      source  = "terraform-redhat/rhcs"
      version = ">= 1.7.0"
    }
  }
}

provider "google" {
  region = var.region
}

# OCM credentials. Set via env var (RHCS_TOKEN / RHCS_CLIENT_ID / RHCS_CLIENT_SECRET)
# or fill in here. client_id + client_secret is preferred for production use
# (offline tokens expire in ~15 minutes which is shorter than the cluster create).
provider "rhcs" {
  # token         = var.ocm_offline_token
  # client_id     = var.ocm_client_id
  # client_secret = var.ocm_client_secret
}
