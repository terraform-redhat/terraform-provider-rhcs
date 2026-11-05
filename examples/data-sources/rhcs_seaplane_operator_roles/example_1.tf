# Discover an existing V2 operator role set (e.g. one created by the rosa
# CLI) for import into a rhcs_seaplane_operator_roles resource:
#
#   terraform import rhcs_seaplane_operator_roles.roles my-cluster
data "rhcs_seaplane_operator_roles" "roles" {
  operator_roles_prefix = "my-cluster"
}

output "role_arns" {
  value = { for role in data.rhcs_seaplane_operator_roles.roles.roles : role.suffix => role.arn }
}
