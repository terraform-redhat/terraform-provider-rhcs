terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

module rosa_operator_roles {
    source = "./operator_roles"
    count = 6

    cluster_id = var.cluster_id
    rh_oidc_provider_url = var.rh_oidc_provider_url
    rh_oidc_provider_thumbprint = var.rh_oidc_provider_thumbprint
    operator_role_properties = var.operator_roles_properties[count.index]
}

