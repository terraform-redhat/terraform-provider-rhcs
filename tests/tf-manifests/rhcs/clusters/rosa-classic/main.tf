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
locals {
  versionfilter = var.openshift_version == null ? "" : " and id like '%${var.openshift_version}%'"
}

data "rhcs_versions" "version" {
  search = "enabled='t' and rosa_enabled='t' and channel_group='${var.channel_group}'${local.versionfilter}"
  # order = "id desc"

}
locals {
  version = data.rhcs_versions.version.items[0].name
}



# data "aws_subnet" "aws_subnets"{
#   for_each = toset(var.aws_subnet_ids)
#   id = each.key

# }

locals {
  # aws_subnet_ids = var.private_link?module.vpc[0].private_subnets:concat(module.vpc[0].private_subnets,module.vpc[0].private_subnets)
  aws_subnet_ids = var.aws_subnet_ids
}
// ************** Managed oidc ***************
resource "rhcs_rosa_oidc_config" "oidc_config_managed" {
  count   = var.oidc_config == "managed" ? 1 : 0
  managed = true
}


// ************** UnManaged oidc **************
# Generates the OIDC config resources' names
resource "rhcs_rosa_oidc_config_input" "oidc_input" {
  count  = var.oidc_config == "un-managed" ? 1 : 0
  region = var.aws_region
}

# Create the OIDC config resources on AWS
module "oidc_config_input_resources" {
  count   = var.oidc_config == "un-managed" ? 1 : 0
  source  = "terraform-redhat/rosa-sts/aws"
  version = ">= 0.0.12"

  create_oidc_config_resources = var.oidc_config == "un-managed"

  bucket_name             = rhcs_rosa_oidc_config_input.oidc_input[0].bucket_name
  discovery_doc           = rhcs_rosa_oidc_config_input.oidc_input[0].discovery_doc
  jwks                    = rhcs_rosa_oidc_config_input.oidc_input[0].jwks
  private_key             = rhcs_rosa_oidc_config_input.oidc_input[0].private_key
  private_key_file_name   = rhcs_rosa_oidc_config_input.oidc_input[0].private_key_file_name
  private_key_secret_name = rhcs_rosa_oidc_config_input.oidc_input[0].private_key_secret_name
}

# Create unmanaged OIDC config
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

module "operator_roles_and_oidc_provider" {
  count   = var.oidc_config == null ? 0 : 1
  source  = "terraform-redhat/rosa-sts/aws"
  version = ">= 0.0.12"

  create_operator_roles = true
  create_oidc_provider  = true

  cluster_id                  = ""
  rh_oidc_provider_thumbprint = var.oidc_config == "managed" ? rhcs_rosa_oidc_config.oidc_config_managed[0].thumbprint : rhcs_rosa_oidc_config.oidc_config_unmanaged[0].thumbprint
  rh_oidc_provider_url        = var.oidc_config == "managed" ? rhcs_rosa_oidc_config.oidc_config_managed[0].oidc_endpoint_url : rhcs_rosa_oidc_config.oidc_config_unmanaged[0].oidc_endpoint_url
  operator_roles_properties   = data.rhcs_rosa_operator_roles.operator_roles.operator_iam_roles
}
locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Installer-Role",
    support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.account_role_prefix}-Worker-Role"
    },
    operator_role_prefix = var.operator_role_prefix,
    oidc_config_id       = var.oidc_config == "managed" ? rhcs_rosa_oidc_config.oidc_config_managed[0].id : rhcs_rosa_oidc_config.oidc_config_unmanaged[0].id
  }
}
/* locals {
  oidc_config_map = {
    "":null,
    managed: rhcs_rosa_oidc_config.oidc_config_managed[0].id,
    unmanaged: rhcs_rosa_oidc_config.oidc_config_unmanaged[0].id
  }
} */

/* locals{
  local.sts_roles[]
} */

data "aws_caller_identity" "current" {
}
// **********
resource "rhcs_cluster_rosa_classic" "rosa_sts_cluster" {
  name               = var.cluster_name
  version            = local.version
  channel_group      = var.channel_group
  cloud_region       = var.aws_region
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = var.aws_availability_zones
  multi_az           = var.multi_az
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts                         = local.sts_roles
  replicas                    = var.replicas
  proxy                       = var.proxy
  autoscaling_enabled         = var.autoscaling.autoscaling_enabled
  min_replicas                = var.autoscaling.min_replicas
  max_replicas                = var.autoscaling.max_replicas
  ec2_metadata_http_tokens    = var.aws_http_tokens_state
  aws_private_link            = var.private_link
  private                     = var.private
  aws_subnet_ids              = local.aws_subnet_ids
  compute_machine_type        = var.compute_machine_type
  default_mp_labels           = var.default_mp_labels
  disable_scp_checks          = var.disable_scp_checks
  disable_workload_monitoring = var.disable_workload_monitoring
  etcd_encryption             = var.etcd_encryption
  fips                        = var.fips
  host_prefix                 = var.host_prefix
  kms_key_arn                 = var.kms_key_arn
  machine_cidr                = var.machine_cidr
  service_cidr                = var.service_cidr
  pod_cidr                    = var.pod_cidr
  tags                        = var.tags
  destroy_timeout             = 120


  # depends_on = [
  #   module.vpc
  #   ]
}


resource "rhcs_cluster_wait" "rosa_cluster" {
  cluster = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  timeout = 120
}


module "operator_roles" {
  source  = "terraform-redhat/rosa-sts/aws"
  version = "0.0.11"

  create_operator_roles = true
  create_oidc_provider  = var.oidc_config == null ? true : false
  create_account_roles  = false

  cluster_id                  = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
  rh_oidc_provider_thumbprint = rhcs_cluster_rosa_classic.rosa_sts_cluster.sts.thumbprint
  rh_oidc_provider_url        = rhcs_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url
  operator_roles_properties   = data.rhcs_rosa_operator_roles.operator_roles.operator_iam_roles
}

