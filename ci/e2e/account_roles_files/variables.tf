variable ocm_environment {
    type = string
    default = "staging"
}

variable token {
    type = string
}

variable url {
    type = string
}

variable aws_access_key {
    type = string
}

variable aws_secret_key {
    type = string
}

variable aws_region {
    type = string
    default = "us-east-1"
}

variable openshift_version {
    type = string
    default = "4.13"
}

variable account_role_prefix {
    type = string
}
