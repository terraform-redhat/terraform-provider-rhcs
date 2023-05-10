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

variable aws_region {
    type = string
    default = "us-east-1"
}

variable aws_availability_zones {
    type = string
    default = "us-east-1a"
}

variable replicas {
    type = string
    default = "3"
}
