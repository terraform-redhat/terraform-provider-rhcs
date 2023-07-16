variable token {
  type = string
  sensitive = true
}

variable url {
    type = string
    default = "https://api.stage.openshift.com"
}

variable operator_role_prefix {
    type = string
}

variable account_role_prefix {
    type = string
}

variable cluster_name {
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

variable aws_availability_zones {
    type      = list(string)
    default = ["us-east-1a"]
}

variable replicas {
    type = string
    default = "3"
}

variable openshift_version {
    type = string
    default = "4.13.0"
}

variable channel_group {
    type = string
    default = "stable"
}