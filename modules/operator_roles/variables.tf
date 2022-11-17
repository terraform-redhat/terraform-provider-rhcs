variable cluster_id {
    description = "cluster ID"
    type = string
}

variable number_of_roles {
    description = "number of roles"
    type = number
}

variable "operator_roles_properties" {
  description = "operator role properties"
  type = list(object({
    operator_name     = string
    role_name = string
    namespace = string
    role_arn = string
    policy_name = string
    service_accounts = list(string)
  }))
}


variable rh_oidc_provider_url {
    description = "oidc provider url"
    type = string
    default = "rh-oidc.s3.us-east-1.amazonaws.com"
}

variable rh_oidc_provider_thumbprint {
    description = "Thumbprint for https://rh-oidc.s3.us-east-1.amazonaws.com"
    type = string
    default = "917e732d330f9a12404f73d8bea36948b929dffc"
}


variable account_role_prefix {
    description = "account role prefix"
    type = string
    default = "ManagedOpenShift"
}

