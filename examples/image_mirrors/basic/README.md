# Basic Image Mirror Example

This example demonstrates how to create a basic image mirror configuration for a ROSA HCP cluster.

## What This Example Does

- Creates a single image mirror for nginx from Docker Hub to Quay.io
- Shows how to use variables for flexibility
- Demonstrates retrieving all image mirrors for a cluster using the data source

## Prerequisites

- An existing ROSA HCP cluster
- Terraform installed
- RHCS provider configured

## Usage

1. Set the required variables:
   ```bash
   export TF_VAR_cluster_id="your-cluster-id"
   ```

2. Optionally customize the source and mirrors:
   ```bash
   export TF_VAR_source_registry="docker.io/library/nginx"
   export TF_VAR_mirrors='["quay.io/my-org/nginx"]'
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| cluster_id | The ID of the ROSA HCP cluster | string | n/a | yes |
| source_registry | The source registry to mirror | string | "docker.io/library/nginx" | no |
| mirrors | List of mirror registries | list(string) | ["quay.io/my-org/nginx"] | no |
| type | The type of mirror | string | "digest" | no |

## Outputs

| Name | Description |
|------|-------------|
| image_mirror_id | The unique identifier of the created image mirror |
| image_mirror_source | The source registry being mirrored |
| image_mirror_mirrors | List of configured mirror registries |
| image_mirror_type | The type of the image mirror |
| creation_timestamp | When the image mirror was created |
| last_update_timestamp | When the image mirror was last updated |
| all_cluster_mirrors | All image mirrors configured for the cluster |

## Example Output

```
image_mirror_id = "mirror-123456"
image_mirror_source = "docker.io/library/nginx"
image_mirror_mirrors = ["quay.io/my-org/nginx"]
image_mirror_type = "digest"
creation_timestamp = "2024-01-15T10:30:00Z"
last_update_timestamp = "2024-01-15T10:30:00Z"
```