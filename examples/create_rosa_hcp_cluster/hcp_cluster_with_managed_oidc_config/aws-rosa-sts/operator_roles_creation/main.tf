terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

module "rosa_operator_roles" {
  source = "./operator_roles"
  count  = 8

  cluster_id               = var.cluster_id
  rh_oidc_provider_url     = var.rh_oidc_provider_url
  operator_role_properties = var.operator_roles_properties[count.index]
  permissions_boundary     = var.permissions_boundary
  tags                     = var.tags
  path                     = var.path
}
