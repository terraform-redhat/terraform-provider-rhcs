---
page_title: "Image Mirrors for ROSA HCP Clusters"
subcategory: ""
description: |-
  Guide for configuring image mirrors in ROSA HCP clusters using the RHCS Terraform provider.
---

# Image Mirrors for ROSA HCP Clusters

**Important**: Image mirrors are only supported for IDMS (ImageDigestMirrorSet) through ICSP (ImageContentSourcePolicy). This feature configures registry overrides that work with HyperShift's IDMS/ICSP detection mechanism.

For detailed information on setting up IDMS/ICSP for management clusters, see: [IDMS/ICSP for Management Clusters](https://hypershift-docs.netlify.app/how-to/disconnected/idms-icsp-for-management-clusters/)

Image mirrors provide a way to configure alternative registry endpoints for container images in ROSA HCP (Hosted Control Plane) clusters. This feature ensures that your cluster can pull images from mirror registries when the original source registry is unavailable or when you want to use a local registry for performance or security reasons.

## Overview

The RHCS Terraform provider offers two main resources for managing image mirrors:

- **`rhcs_image_mirror`** - Creates and manages individual image mirror configurations
- **`rhcs_image_mirrors`** (data source) - Retrieves all image mirrors configured for a cluster

## Prerequisites

- A ROSA HCP cluster (image mirrors are not supported on Classic clusters)
- Appropriate IAM permissions to manage cluster configurations
- Access to the mirror registries you want to configure

## Basic Usage

### Creating a Simple Image Mirror

```terraform
resource "rhcs_image_mirror" "nginx_mirror" {
  cluster_id = "your-cluster-id"
  source     = "docker.io/library/nginx"
  mirrors    = ["quay.io/my-org/nginx"]
  type       = "digest"  # Optional, defaults to "digest"
}
```

### Retrieving Existing Image Mirrors

```terraform
data "rhcs_image_mirrors" "cluster_mirrors" {
  cluster_id = "your-cluster-id"
}

output "all_mirrors" {
  value = data.rhcs_image_mirrors.cluster_mirrors.image_mirrors
}
```

## Advanced Configuration

### Multiple Mirror Registries

You can configure multiple mirror registries for a single source. The cluster will attempt to pull from mirrors in the order they are specified:

```terraform
resource "rhcs_image_mirror" "multi_mirror" {
  cluster_id = var.cluster_id
  source     = "docker.io/library/alpine"
  mirrors    = [
    "quay.io/my-org/alpine",           # Primary mirror
    "registry.example.com/alpine",     # Secondary mirror
    "docker.io/my-backup/alpine"       # Tertiary mirror
  ]
}
```

### Using Variables for Flexibility

```terraform
variable "cluster_id" {
  type        = string
  description = "The ROSA HCP cluster ID"
}

variable "registry_mirrors" {
  type = map(list(string))
  description = "Map of source registries to their mirror lists"
  default = {
    "docker.io/library/nginx" = ["quay.io/my-org/nginx"]
    "docker.io/library/redis" = ["registry.example.com/redis", "quay.io/backup/redis"]
  }
}

resource "rhcs_image_mirror" "configured_mirrors" {
  for_each = var.registry_mirrors

  cluster_id = var.cluster_id
  source     = each.key
  mirrors    = each.value
}
```

## Complete Example with Cluster Creation

```terraform
terraform {
  required_providers {
    rhcs = {
      version = ">= 1.1.0"
      source  = "terraform-redhat/rhcs"
    }
  }
}

locals {
  sts_roles = {
    role_arn         = "arn:aws:iam::123456789012:role/my-HCP-ROSA-Installer-Role"
    support_role_arn = "arn:aws:iam::123456789012:role/my-HCP-ROSA-Support-Role"
    instance_iam_roles = {
      worker_role_arn = "arn:aws:iam::123456789012:role/my-HCP-ROSA-Worker-Role"
    }
    operator_role_prefix = "my-operator"
    oidc_config_id       = "my-oidc-config-id"
  }
}

# Create HCP cluster
resource "rhcs_cluster_rosa_hcp" "example_cluster" {
  name                   = "example-hcp-cluster"
  cloud_region           = "us-east-1"
  aws_account_id         = "123456789012"
  aws_billing_account_id = "123456789012"
  aws_subnet_ids         = ["subnet-123", "subnet-456"]
  availability_zones     = ["us-east-1a", "us-east-1b"]
  replicas               = 2
  version                = "4.15.9"

  sts                                 = local.sts_roles
  wait_for_create_complete            = true
  wait_for_std_compute_nodes_complete = true
}

# Configure image mirrors after cluster creation
resource "rhcs_image_mirror" "nginx_mirror" {
  cluster_id = rhcs_cluster_rosa_hcp.example_cluster.id
  source     = "docker.io/library/nginx"
  mirrors    = ["quay.io/my-org/nginx"]

  depends_on = [rhcs_cluster_rosa_hcp.example_cluster]
}

resource "rhcs_image_mirror" "redis_mirror" {
  cluster_id = rhcs_cluster_rosa_hcp.example_cluster.id
  source     = "docker.io/library/redis"
  mirrors    = [
    "registry.example.com/redis",
    "quay.io/backup/redis"
  ]

  depends_on = [rhcs_cluster_rosa_hcp.example_cluster]
}
```

## Important Considerations

### Immutable Fields

- **`cluster_id`**: Cannot be changed after creation (triggers replacement)
- **`source`**: Cannot be changed after creation (triggers replacement)

### Updatable Fields

- **`mirrors`**: Can be modified to add, remove, or reorder mirror registries
- **`type`**: Can be updated (currently only "digest" is supported)

### Behavior Notes

1. **Order Matters**: Mirrors are tried in the order specified in the `mirrors` list
2. **HCP Only**: Image mirrors are only supported on ROSA HCP clusters, not Classic clusters
3. **Registry Authentication**: Ensure your cluster has appropriate pull secrets for accessing mirror registries
4. **Digest Type**: Currently only "digest" type mirrors are supported

## Common Use Cases

### Corporate Registry Mirror

```terraform
# Mirror public images through corporate registry
resource "rhcs_image_mirror" "corporate_mirrors" {
  for_each = {
    "docker.io/library/nginx"     = ["registry.corp.example.com/docker/nginx"]
    "docker.io/library/postgres"  = ["registry.corp.example.com/docker/postgres"]
    "quay.io/prometheus/prometheus" = ["registry.corp.example.com/quay/prometheus"]
  }

  cluster_id = var.cluster_id
  source     = each.key
  mirrors    = each.value
}
```

### High Availability Setup

```terraform
# Multiple mirrors for high availability
resource "rhcs_image_mirror" "ha_mirror" {
  cluster_id = var.cluster_id
  source     = "docker.io/library/mysql"
  mirrors    = [
    "primary-registry.example.com/mysql",    # Primary
    "backup-registry.example.com/mysql",     # Backup
    "quay.io/my-org/mysql"                   # Emergency fallback
  ]
}
```

## Troubleshooting

### Common Issues

1. **Cluster Type Error**: Ensure you're using a ROSA HCP cluster, not a Classic cluster
2. **Registry Access**: Verify that your cluster has pull secrets for accessing mirror registries
3. **Source Registry Format**: Use the full registry path (e.g., "docker.io/library/nginx", not just "nginx")
4. **Mirror Availability**: Test that your mirror registries are accessible from your cluster's network

### Validation

Use the data source to verify your configuration:

```terraform
data "rhcs_image_mirrors" "validation" {
  cluster_id = var.cluster_id
}

output "configured_mirrors" {
  value = {
    for mirror in data.rhcs_image_mirrors.validation.image_mirrors :
    mirror.source => mirror.mirrors
  }
}
```

## Additional Resources

- [ROSA Documentation](https://docs.openshift.com/rosa/)
- [OpenShift Image Registry Configuration](https://docs.openshift.com/container-platform/latest/openshift_images/image-configuration.html)
- [Container Registry Security Best Practices](https://docs.openshift.com/container-platform/latest/security/container_security/security-container-content.html)