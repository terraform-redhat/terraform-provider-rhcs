# ROSA HCP Image Mirrors Examples

This directory contains examples showing how to configure image mirrors for ROSA HCP (Hosted Control Plane) clusters using the RHCS Terraform provider.

Image mirrors allow you to configure alternative registry endpoints for container images, ensuring your cluster can pull images from mirror registries when the original source is unavailable or when you want to use local registries for performance or security reasons.

## Prerequisites

- A ROSA HCP cluster (image mirrors are not supported on Classic clusters)
- Appropriate IAM permissions to manage cluster configurations
- Access to the mirror registries you want to configure

## Examples Included

### 1. Basic Image Mirror (`basic/`)
A simple example showing how to create a single image mirror for an existing ROSA HCP cluster.

### 2. Multiple Mirrors (`multiple_mirrors/`)
Example demonstrating how to configure multiple mirror registries for different source registries.

### 3. Complete Setup (`complete_setup/`)
A comprehensive example that creates a ROSA HCP cluster and configures multiple image mirrors.

## Important Notes

- **HCP Only**: Image mirrors are only supported on ROSA HCP clusters, not Classic clusters
- **Immutable Fields**: `cluster_id` and `source` cannot be changed after creation
- **Order Matters**: Mirrors are tried in the order specified in the `mirrors` list
- **Registry Authentication**: Ensure your cluster has appropriate pull secrets for accessing mirror registries

## Usage

1. Navigate to the desired example directory
2. Copy the example files to your working directory
3. Modify the variables according to your environment
4. Run `terraform init`
5. Run `terraform plan`
6. Run `terraform apply`

## Additional Resources

- [Image Mirrors Guide](../../docs/guides/image-mirrors.md)
- [ROSA Documentation](https://docs.openshift.com/rosa/)
- [OpenShift Image Registry Configuration](https://docs.openshift.com/container-platform/latest/openshift_images/image-configuration.html)