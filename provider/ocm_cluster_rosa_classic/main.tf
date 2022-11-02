terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

module operator_role {
    source = "./operator_roles"
    cluster_id = var.cluster_id
    operator_role_prefix = var.operator_role_prefix
    rh_oidc_provider_thumbprint = var.rh_oidc_provider_thumbprint
    rh_oidc_provider_url = var.rh_oidc_provider_url
}
