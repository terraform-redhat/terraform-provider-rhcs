resource "aws_iam_role" "csi_drivers_role" {
  name = "${var.operator_role_prefix}-openshift-cluster-csi-drivers-ebs-cloud-cre"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Condition = {
            StringEquals = {
                "${var.rh_oidc_provider_url}:sub" = [
                  "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
                  "system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa"
                  ]
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
    operator_namespace = "openshift-cluster-csi-drivers"
    operator_name = "ebs-cloud-credentials"
  }
}

resource "aws_iam_role_policy_attachment" "csi_drivers_role_policy_attachment" {
  role = aws_iam_role.csi_drivers_role.name
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/ManagedOpenShift-openshift-cluster-csi-drivers-ebs-cloud-credent"
}