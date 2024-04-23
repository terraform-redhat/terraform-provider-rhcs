variable "url" {
  type    = string
  default = "https://api.stage.openshift.com"
}

variable "cluster" {
  type = string
}

variable "listening_method" {
  type    = string
  default = null
}
