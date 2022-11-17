data "aws_caller_identity" "current" {}

resource "aws_iam_role" "operator_role" {
  #count = length(var.operator_roles_properties)
  count = var.number_of_roles

  name = var.operator_roles_properties[count.index].role_name
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Condition = {
            StringEquals = {
                "${var.rh_oidc_provider_url}:sub" = var.operator_roles_properties[count.index].service_accounts
            }
        }
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${var.rh_oidc_provider_url}"
        }
      },
    ]
  })

  tags = {
    red-hat-managed = true
    rosa_cluster_id = var.cluster_id
    operator_namespace = var.operator_roles_properties[count.index].namespace
    operator_name = var.operator_roles_properties[count.index].operator_name
  }
}

resource "aws_iam_role_policy_attachment" "operator_role_policy_attachment" {
  #count = length(var.operator_roles_properties)
  count = var.number_of_roles

  role = var.operator_roles_properties[count.index].role_name
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/${var.operator_roles_properties[count.index].policy_name}"
}

