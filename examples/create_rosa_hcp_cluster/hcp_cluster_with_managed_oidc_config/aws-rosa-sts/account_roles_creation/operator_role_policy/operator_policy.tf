resource "aws_iam_policy" "operator-policy" {
  name   = var.operator_role_policy_properties.policy_name
  policy = var.operator_role_policy_properties.policy_details

  tags = merge(var.tags, {
    rosa_openshift_version = "${var.rosa_openshift_version}"
    rosa_role_prefix       = "${var.account_role_prefix}"
    operator_namespace     = "${var.operator_role_policy_properties.namespace}"
    operator_name          = "${var.operator_role_policy_properties.operator_name}"
  })
}
