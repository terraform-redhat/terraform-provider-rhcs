variable token {
  type = string
  sensitive = true
}

variable operator_role_prefix {
    type = string
}

variable installer_role_arn {
    type = string
}

variable support_role_arn {
    type = string
}

variable control_plane_role_arn {
    type = string
}

variable worker_role_arn {
    type = string
}

variable url {
    type = string
    default = "https://api.openshift.com"
}

variable oidc_endpoint_url {
    type = string
}

variable oidc_private_key_secret_arn {
    type = string
}
