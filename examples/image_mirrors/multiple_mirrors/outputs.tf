output "created_mirrors" {
  description = "Details of all created image mirrors"
  value = {
    for source, mirror in rhcs_image_mirror.mirrors : source => {
      id                    = mirror.id
      source                = mirror.source
      mirrors               = mirror.mirrors
      type                  = mirror.type
      creation_timestamp    = mirror.creation_timestamp
      last_update_timestamp = mirror.last_update_timestamp
    }
  }
}

output "mirror_count" {
  description = "Total number of image mirrors created"
  value       = length(rhcs_image_mirror.mirrors)
}

output "sources_configured" {
  description = "List of source registries that have mirrors configured"
  value       = keys(rhcs_image_mirror.mirrors)
}

output "all_cluster_mirrors" {
  description = "All image mirrors retrieved from the cluster"
  value       = data.rhcs_image_mirrors.cluster_mirrors.image_mirrors
}

output "mirror_summary" {
  description = "Summary of mirror configurations"
  value = {
    for source, mirror in rhcs_image_mirror.mirrors : source => {
      mirror_count   = length(mirror.mirrors)
      primary_mirror = mirror.mirrors[0]
    }
  }
}