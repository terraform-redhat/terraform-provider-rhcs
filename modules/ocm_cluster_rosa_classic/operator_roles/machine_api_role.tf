resource "aws_iam_role" "machine_api_role" {
  name = "${var.operator_role_prefix}-openshift-machine-api-aws-cloud-credentials"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Condition = {
            StringEquals = {
                "${var.rh_oidc_provider_url}:sub" = ["system:serviceaccount:openshift-machine-api:machine-api-controllers"]
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
    operator_namespace = "openshift-machine-api"
    operator_name = "aws-cloud-credentials"
  }
}

resource "aws_iam_role_policy_attachment" "machine_api_role_policy_attachment" {
  role = aws_iam_role.machine_api_role.name
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/ManagedOpenShift-openshift-machine-api-aws-cloud-credentials"
}