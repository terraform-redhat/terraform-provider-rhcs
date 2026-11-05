data "aws_caller_identity" "current" {}

provider "rhcs" {
  seaplane_url   = var.seaplane_url
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_caller_arn = data.aws_caller_identity.current.arn
  aws_region     = var.aws_region
}

resource "rhcs_seaplane_oidc_config" "oidc" {
  managed = true
}

# Creates the full V2 operator role set for the prefix, including the worker
# instance role and its EC2 instance profile. Trust policies are generated
# from the shared seaplane iam definitions — no hand-written IAM JSON.
resource "rhcs_seaplane_operator_roles" "roles" {
  operator_roles_prefix = "my-cluster"
  oidc_issuer_url       = rhcs_seaplane_oidc_config.oidc.issuer_url
}

variable "seaplane_url" {
  type        = string
  description = "Seaplane Platform API endpoint."
}

variable "aws_region" {
  type        = string
  description = "AWS region used to sign Seaplane Platform API requests."
}
