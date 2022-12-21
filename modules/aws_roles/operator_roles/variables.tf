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
    description = "Thumbprint for the variable `rh_oidc_provider_url`"
    type = string
    default = "917e732d330f9a12404f73d8bea36948b929dffc"
}

variable operator_role_properties {
    description = ""
    type = object({
        role_name = string
        policy_name = string
        service_accounts = list(string***REMOVED***
        operator_name = string
        operator_namespace = string
    }***REMOVED***

}
