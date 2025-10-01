output "image_mirror_id" {
  description = "The unique identifier of the created image mirror"
  value       = rhcs_image_mirror.nginx_mirror.id
}

output "image_mirror_source" {
  description = "The source registry that is being mirrored"
  value       = rhcs_image_mirror.nginx_mirror.source
}

output "image_mirror_mirrors" {
  description = "List of mirror registries configured"
  value       = rhcs_image_mirror.nginx_mirror.mirrors
}

output "image_mirror_type" {
  description = "The type of the image mirror"
  value       = rhcs_image_mirror.nginx_mirror.type
}

output "creation_timestamp" {
  description = "Timestamp when the image mirror was created"
  value       = rhcs_image_mirror.nginx_mirror.creation_timestamp
}

output "last_update_timestamp" {
  description = "Timestamp when the image mirror was last updated"
  value       = rhcs_image_mirror.nginx_mirror.last_update_timestamp
}

output "all_cluster_mirrors" {
  description = "All image mirrors configured for the cluster"
  value       = data.rhcs_image_mirrors.cluster_mirrors.image_mirrors
}