output operator_iam_roles {
  value =  data.ocm_rosa_operator_roles.operator_roles.operator_iam_roles
}

output cluster_id {
  value = ocm_cluster_rosa_classic.rosa_sts_cluster.id
}

output rh_oidc_provider_thumbprint {
  value = ocm_cluster_rosa_classic.rosa_sts_cluster.sts.thumbprint
}

output rh_oidc_provider_url {
   value = ocm_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url
}