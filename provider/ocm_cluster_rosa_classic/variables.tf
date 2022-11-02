variable cluster_id {
    description = "cluster ID"
    type = string
}

variable operator_role_prefix {
    description = "operator role prefix"
    type = string
    default = "terraform-rosa-classic"
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
