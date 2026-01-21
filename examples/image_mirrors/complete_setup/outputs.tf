# Cluster Outputs
output "cluster_id" {
  description = "ID of the created ROSA HCP cluster"
  value       = rhcs_cluster_rosa_hcp.cluster.id
}

output "cluster_name" {
  description = "Name of the created cluster"
  value       = rhcs_cluster_rosa_hcp.cluster.name
}

output "cluster_state" {
  description = "State of the cluster"
  value       = rhcs_cluster_rosa_hcp.cluster.state
}

output "cluster_api_url" {
  description = "API URL of the cluster"
  value       = rhcs_cluster_rosa_hcp.cluster.api_url
}

output "cluster_console_url" {
  description = "Console URL of the cluster"
  value       = rhcs_cluster_rosa_hcp.cluster.console_url
}

output "cluster_domain" {
  description = "Domain of the cluster"
  value       = rhcs_cluster_rosa_hcp.cluster.domain
}

# Image Mirror Outputs
output "configured_image_mirrors" {
  description = "Details of all configured image mirrors"
  value = {
    for source, mirror in rhcs_image_mirror.cluster_mirrors : source => {
      id                    = mirror.id
      source                = mirror.source
      mirrors               = mirror.mirrors
      type                  = mirror.type
      creation_timestamp    = mirror.creation_timestamp
      last_update_timestamp = mirror.last_update_timestamp
    }
  }
}

output "image_mirror_count" {
  description = "Total number of image mirrors configured"
  value       = length(rhcs_image_mirror.cluster_mirrors)
}

output "mirrored_sources" {
  description = "List of source registries that have mirrors configured"
  value       = keys(rhcs_image_mirror.cluster_mirrors)
}

output "all_cluster_mirrors" {
  description = "All image mirrors retrieved from the cluster (includes any existing ones)"
  value       = data.rhcs_image_mirrors.all_mirrors.image_mirrors
}

# Summary Output
output "deployment_summary" {
  description = "Summary of the complete deployment"
  value = {
    cluster = {
      id     = rhcs_cluster_rosa_hcp.cluster.id
      name   = rhcs_cluster_rosa_hcp.cluster.name
      state  = rhcs_cluster_rosa_hcp.cluster.state
      region = rhcs_cluster_rosa_hcp.cluster.cloud_region
    }
    image_mirrors = {
      count   = length(rhcs_image_mirror.cluster_mirrors)
      sources = keys(rhcs_image_mirror.cluster_mirrors)
    }
  }
}