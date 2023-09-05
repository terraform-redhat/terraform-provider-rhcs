variable "cluster_id" {
  description = "cluster ID"
  type        = string
  default     = ""
}

variable "permissions_boundary" {
  description = "The ARN of the policy that is used to set the permissions boundary for the IAM roles in STS clusters."
  type        = string
  default     = ""
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}

# ******************* operator roles variables
variable "rh_oidc_provider_url" {
  description = "oidc provider url"
  type        = string
  default     = "rh-oidc.s3.us-east-1.amazonaws.com"
}

variable "operator_roles_properties" {
  description = "List of ROSA Operator IAM Roles"
  type = list(object({
    role_name          = string
    policy_name        = string
    service_accounts   = list(string)
    operator_name      = string
    operator_namespace = string
  }))
  default = []
}

variable "create_operator_roles" {
  description = "When using OIDC Config and reusing the operator roles set to false so as not to create operator roles"
  type        = bool
  default     = false
}

variable "create_oidc_provider" {
  description = "When using OIDC Config and reusing the OIDC provider set to false so as not to create identity provider"
  type        = bool
  default     = false
}

# ******************* account roles variables
variable "create_account_roles" {
  description = "This attribute determines whether the module should create account roles or not"
  type        = bool
  default     = false
}

variable "rh_oidc_provider_thumbprint" {
  description = "Thumbprint for https://rh-oidc.s3.us-east-1.amazonaws.com"
  type        = string
  default     = "917e732d330f9a12404f73d8bea36948b929dffc"
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "rosa_openshift_version" {
  description = "Desired version of OpenShift for the cluster, for example '4.1.0'. If version is greater than the currently running version, an upgrade will be scheduled."
  type        = string
  default     = ""
}

variable "ocm_environment" {
  description = "The OCM environments should be one of those: production, staging, integration, local"
  type        = string
  default     = ""
}

variable "account_role_policies" {
  description = "account role policies details for account roles creation"
  type = object({
    sts_installer_permission_policy             = string
    sts_support_permission_policy               = string
    sts_instance_worker_permission_policy       = string
    sts_instance_controlplane_permission_policy = string
  })
  default = null
}

variable "operator_role_policies" {
  description = "operator role policies details for operator roles creation"
  type = object({
    openshift_cloud_network_config_controller_cloud_credentials_policy                = string
    openshift_cluster_csi_drivers_ebs_cloud_credentials_policy                        = string
    openshift_image_registry_installer_cloud_credentials_policy                       = string
    openshift_ingress_operator_cloud_credentials_policy                               = string
    openshift_capa_controller_manager_credentials_policy                              = string
    openshift_control_plane_operator_credentials_policy                               = string
    openshift_kms_provider_credentials_policy                                         = string
    openshift_kube_controller_manager_credentials_policy                              = string
  })
  default = null
}

variable "all_versions" {
  description = "OpenShift versions"
  type        = object({
    item = object({
      id   = string
      name = string
    })
    search = string
    order  = string
    items  = list(object({
      id   = string
      name = string
    }))
  })
  default = null
}

# ******************* OIDC config resources
variable "create_oidc_config_resources" {
  description = "This attribute determines whether the module should create OIDC config resources"
  type        = bool
  default     = false
}

variable "bucket_name" {
  description = "The S3 bucket name"
  type        = string
  default     = ""
}

variable "discovery_doc" {
  description = "The discovery document string file"
  type        = string
  default     = ""
}

variable "jwks" {
  description = "Json web key set string file"
  type        = string
  default     = ""
}

variable "private_key" {
  description = "RSA private key"
  type        = string
  default     = ""
}

variable "private_key_file_name" {
  description = "The private key file name"
  type        = string
  default     = ""
}

variable "private_key_secret_name" {
  description = "The secret name that store the private key"
  type        = string
  default     = ""
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}
