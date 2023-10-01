
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

variable "operator_role_prefix" {
  type    = string
  default = ""
}
variable "oidc_config"{
    type    = string
    default = ""
}
variable "aws_region"{
    type    = string
    default = "us-east-2"
}