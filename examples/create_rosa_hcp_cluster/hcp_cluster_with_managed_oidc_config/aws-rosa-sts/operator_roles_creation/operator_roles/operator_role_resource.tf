data "aws_caller_identity" "current" {}

resource "aws_iam_role" "operator_role" {
  name                 = var.operator_role_properties.role_name
  permissions_boundary = var.permissions_boundary
  path                 = var.path
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Condition = {
          StringEquals = {
            "${var.rh_oidc_provider_url}:sub" = var.operator_role_properties.service_accounts
          }
        }
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${var.rh_oidc_provider_url}"
        }
      },
    ]
  })

  tags = merge(var.tags, {
    red-hat-managed    = true
    rosa_cluster_id    = var.cluster_id
    operator_namespace = var.operator_role_properties.operator_namespace
    operator_name      = var.operator_role_properties.operator_name
  })
}

resource "aws_iam_role_policy_attachment" "operator_role_policy_attachment" {
  role       = aws_iam_role.operator_role.name
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/${var.operator_role_properties.policy_name}"
}

