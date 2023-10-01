
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      version = ">= 1.0.1"
      source  = "terraform.local/local/rhcs"
    }
  }
}

provider "rhcs" {
  token = var.token
  url   = var.url
}
provider "aws" {
  region = var.aws_region
}
data "aws_caller_identity" "current" {
}

// ************** Managed oidc config ***************
resource "rhcs_rosa_oidc_config" "oidc_config_managed" {
  count   = var.oidc_config == "managed" ? 1 : 0
  managed = true
}


// ************** UnManaged oidc config **************
# Generates the unmanaged OIDC config resources' names
resource "rhcs_rosa_oidc_config_input" "oidc_input" {
  count  = var.oidc_config == "un-managed" ? 1 : 0
  region = var.aws_region
}

# Create the OIDC config resources on AWS
module "oidc_config_input_resources" {
  count   = var.oidc_config == "un-managed" ? 1 : 0
  source  = "terraform-redhat/rosa-sts/aws"
  version = ">= 0.0.14"

  create_oidc_config_resources = var.oidc_config == "un-managed"

  bucket_name             = rhcs_rosa_oidc_config_input.oidc_input[0].bucket_name
  discovery_doc           = rhcs_rosa_oidc_config_input.oidc_input[0].discovery_doc
  jwks                    = rhcs_rosa_oidc_config_input.oidc_input[0].jwks
  private_key             = rhcs_rosa_oidc_config_input.oidc_input[0].private_key
  private_key_file_name   = rhcs_rosa_oidc_config_input.oidc_input[0].private_key_file_name
  private_key_secret_name = rhcs_rosa_oidc_config_input.oidc_input[0].private_key_secret_name
}

# Create unmanaged OIDC config in OCM
resource "rhcs_rosa_oidc_config" "oidc_config_unmanaged" {
  count              = var.oidc_config == "un-managed" ? 1 : 0
  managed            = false
  secret_arn         = module.oidc_config_input_resources[0].secret_arn
  issuer_url         = rhcs_rosa_oidc_config_input.oidc_input[0].issuer_url
  installer_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Installer-Role"
}

data "rhcs_rosa_operator_roles" "operator_roles" {
  operator_role_prefix = var.operator_role_prefix
  account_role_prefix  = var.account_role_prefix
}


# Create oidc provider and operator roles on AWS
module "operator_roles_and_oidc_provider" {
  count   = var.oidc_config == null ? 0 : 1
  source  = "terraform-redhat/rosa-sts/aws"
  version = ">= 0.0.14"

  create_operator_roles = true
  create_oidc_provider  = true

  cluster_id                  = ""
  rh_oidc_provider_thumbprint = var.oidc_config == "managed" ? rhcs_rosa_oidc_config.oidc_config_managed[0].thumbprint : rhcs_rosa_oidc_config.oidc_config_unmanaged[0].thumbprint
  rh_oidc_provider_url        = var.oidc_config == "managed" ? rhcs_rosa_oidc_config.oidc_config_managed[0].oidc_endpoint_url : rhcs_rosa_oidc_config.oidc_config_unmanaged[0].oidc_endpoint_url
  operator_roles_properties   = data.rhcs_rosa_operator_roles.operator_roles.operator_iam_roles
}