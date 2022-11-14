locals {
    operator_roles = data.ocm_rosa_operator_roles.operator_roles.operator_iam_roles
}



resource "aws_iam_role" "operator_role" {
  depends_on = [data.ocm_rosa_operator_roles.operator_roles]
  count = length(local.operator_roles)

  name = local.operator_roles[count.index].role_name
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Condition = {
            StringEquals = {
                "${ocm_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url}:sub" = local.operator_roles[count.index].service_accounts
            }
        }
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${ocm_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url}"
        }
      },
    ]
  })

  tags = {
    red-hat-managed = true
    rosa_cluster_id = ocm_cluster_rosa_classic.rosa_sts_cluster.id
    operator_namespace = local.operator_roles[count.index].namespace
    operator_name = local.operator_roles[count.index].operator_name
  }
}


resource "aws_iam_role_policy_attachment" "cloud-credential_role_policy_attachment" {
  count = length(local.operator_roles)
  role = aws_iam_role.operator_role[count.index].name
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/${local.operator_roles[count.index].policy_name}"
}


resource "aws_iam_openid_connect_provider" "oidc_provider" {
  url = "https://${ocm_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url}"

  client_id_list = [
    "openshift",
    "sts.amazonaws.com"
  ]

  tags = {
    rosa_cluster_id = ocm_cluster_rosa_classic.rosa_sts_cluster.id
  }
  thumbprint_list = ["${ocm_cluster_rosa_classic.rosa_sts_cluster.sts.thumbprint}"]
}


