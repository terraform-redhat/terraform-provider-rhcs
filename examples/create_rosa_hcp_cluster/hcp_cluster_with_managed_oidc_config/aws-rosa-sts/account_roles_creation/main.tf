terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

resource "random_string" "default_random" {
  length  = 4
  special = false
  upper   = false
}

locals {
  account_role_generated_if_needed = var.account_role_prefix != ""? var.account_role_prefix : "account-role-${random_string.default_random.result}"
}

module "rosa_account_roles" {
  source = "./account_role"
  count  = 2

  account_role_prefix              = local.account_role_generated_if_needed
  rosa_openshift_version           = var.rosa_openshift_version
  account_role_properties          = local.account_roles_properties[count.index]
  instance_account_role_properties = local.instance_account_roles_properties[count.index]
  account_id                       = lookup({ "production" = "710019948333", "staging" = "644306948063", "integration" = "896164604406", "local" = "765374464689" }, var.ocm_environment, "710019948333")
  permissions_boundary             = var.permissions_boundary
  tags                             = var.tags
  path                             = var.path
}

module "rosa_operator_role_policies" {
  source = "./operator_role_policy"
  count  = 8

  account_role_prefix             = local.account_role_generated_if_needed
  rosa_openshift_version          = var.rosa_openshift_version
  operator_role_policy_properties = local.operator_roles_policy_properties[count.index]
  tags                            = var.tags
}

locals {
  account_roles_properties = [{
    # installer
    role_name      = "Installer"
    role_type      = "installer"
    principal      = "RH-Managed-OpenShift-Installer"
    policy_details = var.account_role_policies["sts_installer_permission_policy"]
    },
    {
      # support
      role_name      = "Support"
      role_type      = "support"
      principal      = "RH-Technical-Support-Access"
      policy_details = var.account_role_policies["sts_support_permission_policy"]
  }]

  instance_account_roles_properties = [{
    # worker
    role_name      = "Worker"
    role_type      = "worker"
    policy_details = var.account_role_policies["sts_instance_worker_permission_policy"]
    },
    {
      # control plan
      role_name      = "ControlPlane"
      role_type      = "controlplane"
      policy_details = var.account_role_policies["sts_instance_controlplane_permission_policy"]
  }]

  # TODO: if there is a new policy for a new OCP versions, need to add it here also
  operator_roles_policy_properties = [{
    # openshift-cloud-network-config-controller
    policy_name    = substr("${local.account_role_generated_if_needed}-openshift-cloud-network-config-controller-cloud-credentials", 0, 64)
    policy_details = var.operator_role_policies["openshift_cloud_network_config_controller_cloud_credentials_policy"]
    namespace      = "openshift-cloud-network-config-controller"
    operator_name  = "cloud-redentials"
    },
    {
      # openshift-image-registry
      policy_name    = substr("${local.account_role_generated_if_needed}-openshift-image-registry-installer-cloud-credentials", 0, 64)
      policy_details = var.operator_role_policies["openshift_image_registry_installer_cloud_credentials_policy"]
      namespace      = "openshift-image-registry"
      operator_name  = "installer-cloud-credentials"
    },
    {
      # openshift-ingress-operator
      policy_name    = substr("${local.account_role_generated_if_needed}-openshift-ingress-operator-cloud-credentials", 0, 64)
      policy_details = var.operator_role_policies["openshift_ingress_operator_cloud_credentials_policy"]
      namespace      = "openshift-ingress-operator"
      operator_name  = "cloud-credentials"
    },
    {
      # openshift-cluster-csi-drivers
      policy_name    = substr("${local.account_role_generated_if_needed}-openshift-cluster-csi-drivers-ebs-cloud-credentials", 0, 64)
      policy_details = var.operator_role_policies["openshift_cluster_csi_drivers_ebs_cloud_credentials_policy"]
      namespace      = "openshift-cluster-csi-drivers"
      operator_name  = "ebs-cloud-credentials"
    },
    {
      # capa-controller-manager
      policy_name    = substr("${local.account_role_generated_if_needed}-kube-system-capa-controller-manager", 0, 64)
      policy_details = var.operator_role_policies["openshift_capa_controller_manager_credentials_policy"]
      namespace      = "kube-system"
      operator_name  = "capa-controller-manager"
    },
    {
      # control-plane-operator
      policy_name    = substr("${local.account_role_generated_if_needed}-kube-system-control-plane-operator", 0, 64)
      policy_details = var.operator_role_policies["openshift_control_plane_operator_credentials_policy"]
      namespace      = "kube-system"
      operator_name  = "control-plane-operator"
    },
    {
      # kms-provider
      policy_name    = substr("${local.account_role_generated_if_needed}-kube-system-kms-provider", 0, 64)
      policy_details = var.operator_role_policies["openshift_kms_provider_credentials_policy"]
      namespace      = "kube-system"
      operator_name  = "kms-provider"
    },
    {
      # kube-controller-manager
      policy_name    = substr("${local.account_role_generated_if_needed}-kube-system-kube-controller-manager", 0, 64)
      policy_details = var.operator_role_policies["openshift_kube_controller_manager_credentials_policy"]
      namespace      = "kube-system"
      operator_name  = "kube-controller-manager"
  }]
}
