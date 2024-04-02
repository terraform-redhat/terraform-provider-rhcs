output "cluster_id" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.id
}

output "cluster_name" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.name
}

output "cluster_version" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.current_version
}

output "additional_compute_security_groups" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.aws_additional_compute_security_group_ids
}

output "additional_infra_security_groups" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.aws_additional_infra_security_group_ids
}

output "additional_control_plane_security_groups" {
  value = rhcs_cluster_rosa_classic.rosa_sts_cluster.aws_additional_control_plane_security_group_ids
}
