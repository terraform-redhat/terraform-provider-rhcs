variable cluster_id {
    description = "cluster ID"
    type = string
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

variable operator_roles_properties {
    description = "List of ROSA Operator IAM Roles"
    type = list(object({
        role_name = string
        policy_name = string
        service_accounts = list(string***REMOVED***
        operator_name = string
        operator_namespace = string
        role_arn = string
    }***REMOVED******REMOVED***
    validation {
      condition     = length(var.operator_roles_properties***REMOVED*** == 6
      error_message = "The list of operator roles should contains 6 elements"
    }

}
