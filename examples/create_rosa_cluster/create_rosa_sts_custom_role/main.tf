variable role_prefix {
  type = string  
}

variable role_suffix {
  type = string  
}

variable cluster_name {
  type = string
}

variable token {
  type = string
  sensitive = true
}

variable role_names {
  type = object({
    installer_role = string
    worker_role = string
    support_role = string
    controlplane_role = string
    cloud_credential_role = string
    cloud_network_config_role = string
    csi_driver_role = string
    image_registry_role = string
    ingress_role = string
    machine_api_role = string    
  }
  )
  
  default = {
    installer_role = "ManagedCustom-Installer-Role"
    support_role = "ManagedCustom-Support-Role"
    worker_role = "ManagedCustom-Worker-Role"
    controlplane_role = "ManagedCustom-ControlPlane-Role"
    cloud_credential_role = "custom-cloud-credential-operator-cloud-c"
    cloud_network_config_role = "custom-cloud-network-config-controller-c"
    csi_driver_role = "custom-cluster-csi-drivers-ebs-cloud-cre"
    image_registry_role = "custom-image-registry-installer-cloud-cr"
    ingress_role = "custom-ingress-operator-cloud-credential"
    machine_api_role = "custom-machine-api-aws-cloud-credentials"
  }  
}

terraform {
  required_providers {
    ocm = {
      source = "rh-mobb/ocm"
    }
  }
}

provider "ocm" {
  token = var.token
}

data "aws_caller_identity" "current" {}

locals {
  sts_roles = {
      role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.installer_role}${var.role_suffix}",
      support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.support_role}${var.role_suffix}",
      operator_iam_roles = [
        {
          name =  "cloud-credential-operator-iam-ro-creds",
          namespace = "openshift-cloud-credential-operator",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.cloud_credential_role}${var.role_suffix}",
        },
        {
          name =  "installer-cloud-credentials",
          namespace = "openshift-image-registry",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.image_registry_role}${var.role_suffix}",
        },
        {
          name =  "cloud-credentials",
          namespace = "openshift-ingress-operator",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.ingress_role}${var.role_suffix}",
        },
        {
          name =  "ebs-cloud-credentials",
          namespace = "openshift-cluster-csi-drivers",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.csi_driver_role}${var.role_suffix}",
        },
        {
          name =  "cloud-credentials",
          namespace = "openshift-cloud-network-config-controller",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.cloud_network_config_role}${var.role_suffix}",
        },
        {
          name =  "aws-cloud-credentials",
          namespace = "openshift-machine-api",
          role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.machine_api_role}${var.role_suffix}",
        },
      ]
      instance_iam_roles = {
        master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.controlplane_role}${var.role_suffix}",
        worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.role_prefix}${var.role_names.worker_role}${var.role_suffix}"
      },    
  }
}

resource "ocm_cluster" "rosa_cluster" {
  name           = var.cluster_name
  cloud_provider = "aws"
  cloud_region   = "us-west-2"
  product        = "rosa"
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = ["us-west-2a"]
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  wait = false
  sts = local.sts_roles
}

resource "ocm_cluster_wait" "rosa_cluster" {
  cluster = ocm_cluster.rosa_cluster.id
}

module account_roles {
  source  = "/Users/shading/work/source/terraform-aws-rosa-sts-account-roles"
  role_prefix = var.role_prefix
  role_suffix = var.role_suffix
  controlplane_role = var.role_names.controlplane_role
  installer_role = var.role_names.installer_role
  support_role = var.role_names.support_role
  worker_role = var.role_names.worker_role
}

module operator_roles {
  source  = "/Users/shading/work/source/terraform-aws-rosa-sts-operator-roles"
  operator_role_prefix = var.role_prefix
  operator_role_suffix = var.role_suffix
  cluster_id = ocm_cluster.rosa_cluster.id
  machine_api_policy = module.account_roles.machine_api_policy
  cloud_credential_operator_policy = module.account_roles.cloud_credential_operator_policy
  network_config_controller_policy = module.account_roles.network_config_controller_policy
  image_registry_policy = module.account_roles.image_registry_policy
  ingress_operator_policy = module.account_roles.ingress_operator_policy
  csi_driver_policy = module.account_roles.csi_driver_policy
  cloud_credential_role = var.role_names.cloud_credential_role
  cloud_network_config_role = var.role_names.cloud_network_config_role
  csi_driver_role = var.role_names.csi_driver_role
  ingress_role = var.role_names.ingress_role
  machine_api_role = var.role_names.machine_api_role
  image_registry_role = var.role_names.image_registry_role
}