
locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::account-id-123:role/account-prefix-Installer-Role",
    support_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-Support-Role",
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-ControlPlane-Role",
      worker_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-Worker-Role"
    },
    operator_role_prefix = "operator-prefix",
    oidc_config_id       = "oidc-config-id-123"
  }
}


resource "rhcs_cluster_rosa_classic" "rosa_sts_cluster" {
  name                = "my-cluster"
  cloud_region        = "us-east-2"
  aws_account_id      = "account-id-123"
  availability_zones  = ["us-east-2a"]
  replicas            = 3
  version             = "4.13.12"
  properties = {
    rosa_creator_arn = "aws_caller_identity-current-arn"
  }
  sts = local.sts_roles
  wait_for_create_complete = true
}