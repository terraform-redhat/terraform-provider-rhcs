# Create the AWS IAM OCM role separately with the AWS provider, then
# link its ARN to the OCM organization using this resource.
resource "rhcs_rosa_ocm_role_link" "ocm_role" {
  role_arn = "arn:aws:iam::123456789012:role/ocm-role-ext-org-456"
}
