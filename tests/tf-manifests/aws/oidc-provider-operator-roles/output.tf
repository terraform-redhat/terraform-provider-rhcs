output "operator_role_prefix"{
    value = data.rhcs_rosa_operator_roles.operator_roles.operator_role_prefix
}
output "account_role_prefix"{
    value = data.rhcs_rosa_operator_roles.operator_roles.account_role_prefix
}
output "oidc_config_id"{
    value = var.oidc_config == "managed" ? resource.rhcs_rosa_oidc_config.oidc_config_managed[0].id : rhcs_rosa_oidc_config.oidc_config_unmanaged[0].id
}