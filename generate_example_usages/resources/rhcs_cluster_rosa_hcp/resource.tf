
locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::account-id-123:role/account-prefix-HCP-ROSA-Installer-Role",
    support_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-HCP-ROSA-Support-Role",
    instance_iam_roles = {
      worker_role_arn = "arn:aws:iam::account-id-123:role/account-prefix-HCP-ROSA-Worker-Role"
    },
    operator_role_prefix = "operator-prefix",
    oidc_config_id       = "oidc-config-id-123"
  }
}


resource "rhcs_cluster_rosa_hcp" "rosa_sts_cluster" {
  name                   = "my-cluster"
  cloud_region           = "us-east-2"
  aws_account_id         = "123456789012"
  aws_billing_account_id = "123456789012"
  aws_subnet_ids         = ["subnet-1", "subnet-2"]
  availability_zones     = ["us-west-2a", "us-west-2b"]
  replicas               = 2
  version                = "4.15.9"
  properties = {
    rosa_creator_arn = "aws_caller_identity-current-arn"
  }
  sts                                 = local.sts_roles
  wait_for_create_complete            = true
  wait_for_std_compute_nodes_complete = true
}
