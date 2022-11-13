resource "aws_iam_role" "ingress_operator_role" {
  name = "${var.operator_role_prefix}-openshift-ingress-operator-cloud-credential"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Condition = {
            StringEquals = {
                "${var.rh_oidc_provider_url}:sub" = ["system:serviceaccount:openshift-ingress-operator:ingress-operator"]
            }
        }
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${var.rh_oidc_provider_url}"
        }
      },
    ]
  }***REMOVED***

  tags = {
    red-hat-managed = true
    rosa_cluster_id = var.cluster_id
    operator_namespace = "openshift-ingress-operator"
    operator_name = "cloud-credentials"
  }
}

resource "aws_iam_role_policy_attachment" "ingress_operator_role_policy_attachment" {
  role = aws_iam_role.ingress_operator_role.name
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/${var.account_role_prefix}-openshift-ingress-operator-cloud-credentials"
}