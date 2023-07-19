variable rhcs_environment {
    type = string
    default = "staging"
}

variable openshift_version {
    type = string
    default = "4.13"
}

variable account_role_prefix {
    type = string
    default = ""
}

variable token {
    type = string
}

variable url {
    type = string
    default = "https://api.stage.openshift.com"
}