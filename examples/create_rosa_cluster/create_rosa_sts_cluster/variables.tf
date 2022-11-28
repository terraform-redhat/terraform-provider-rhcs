variable token {
  type = string
  sensitive = true
}

variable operator_role_prefix {
    type = string
}

variable account_role_prefix {
    type = string
    default = ""
}

variable url {
    type = string
    default = "https://api.openshift.com"
}