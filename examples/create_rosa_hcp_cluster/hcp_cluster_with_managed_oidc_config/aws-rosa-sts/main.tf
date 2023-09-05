terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

module "rosa_operator_roles" {
  source = "./operator_roles_creation"
  count  = var.create_operator_roles ? 1 : 0

  cluster_id                = var.cluster_id
  rh_oidc_provider_url      = var.rh_oidc_provider_url
  operator_roles_properties = var.operator_roles_properties
  permissions_boundary      = var.permissions_boundary
  tags                      = var.tags
  path                      = var.path
}

module "rosa_oidc_provider" {
  source = "./oidc_provider_creation"
  count  = var.create_oidc_provider ? 1 : 0

  rh_oidc_provider_url        = var.rh_oidc_provider_url
  rh_oidc_provider_thumbprint = var.rh_oidc_provider_thumbprint
  cluster_id                  = var.cluster_id
  tags                        = var.tags
}


locals {
  selected_version = var.rosa_openshift_version == "" ? "4.13" : var.rosa_openshift_version
  patch_version_list = var.create_account_roles && var.all_versions != null ? [for s in var.all_versions.items : s.name] : []
  minor_version_list = local.patch_version_list != [] ? distinct([for s in local.patch_version_list : format("%s.%s", split(".", s)[0], split(".", s)[1])]) : []
}

# resource "null_resource" "validate_all_version_input" {
#   # validate that all_versions was sent only if the variable "rosa_openshift_version" was set
#   count = var.create_account_roles && var.rosa_openshift_version != "" ? 1 : 0
#   lifecycle {
#     precondition {
#       condition     = var.all_versions != null
#       error_message = "ERROR: Expected `all_versions` as an input variable"
#     }
#   }
# }

# resource "null_resource" "validate_openshift_version" {
#   # validate version only if creating account roles and the variable "rosa_openshift_version" was set and the validation in validate_all_version_input was passed
#   count = var.create_account_roles && var.rosa_openshift_version != "" && null_resource.validate_all_version_input != null ? 1 : 0
#   lifecycle {
#     precondition {
#       condition     = contains(local.minor_version_list, var.rosa_openshift_version)
#       error_message = "ERROR: Expected a valid OpenShift version. Valid versions: ${join(", ",local.minor_version_list)}"
#     }
#   }
# }


module "rosa_account_roles" {
  source = "./account_roles_creation"
  count  = var.create_account_roles ? 1 : 0

  account_role_prefix    = var.account_role_prefix
  rosa_openshift_version = local.selected_version
  ocm_environment        = var.ocm_environment
  account_role_policies  = var.account_role_policies
  operator_role_policies = var.operator_role_policies
  permissions_boundary   = var.permissions_boundary
  tags                   = var.tags
  path                   = var.path
}


module "rosa_oidc_config_resources" {
  source = "./oidc_config_resources"
  count  = var.create_oidc_config_resources ? 1 : 0

  bucket_name             = var.bucket_name
  discovery_doc           = var.discovery_doc
  jwks                    = var.jwks
  private_key             = var.private_key
  private_key_file_name   = var.private_key_file_name
  private_key_secret_name = var.private_key_secret_name
  tags                    = var.tags
}
