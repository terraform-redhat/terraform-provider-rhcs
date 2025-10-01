# Complete ROSA HCP Cluster with Image Mirrors

This example demonstrates a complete setup that creates a ROSA HCP cluster and configures multiple image mirrors in a single Terraform configuration.

## What This Example Does

- Creates a ROSA HCP cluster with proper STS role configuration
- Configures multiple image mirrors after cluster creation
- Provides comprehensive outputs for monitoring and management
- Demonstrates end-to-end setup for a production-ready environment

## Prerequisites

- AWS CLI configured with appropriate permissions
- Terraform installed
- RHCS provider configured
- Pre-created STS roles for ROSA HCP (installer, support, worker roles)
- OIDC configuration setup
- VPC with appropriate subnets

## Required STS Roles

Before running this example, you need to create the following STS roles:

1. **Installer Role**: Manages cluster lifecycle
2. **Support Role**: Provides Red Hat support access
3. **Worker Role**: Used by worker nodes
4. **Operator Roles**: Various operator-specific roles (created separately)

## Usage

1. Create a `terraform.tfvars` file with your configuration:
   ```hcl
   # Basic cluster configuration
   cluster_name = "my-hcp-cluster"
   aws_region = "us-east-1"
   aws_account_id = "123456789012"
   aws_billing_account_id = "123456789012"

   # Network configuration
   subnet_ids = ["subnet-123abc", "subnet-456def"]
   availability_zones = ["us-east-1a", "us-east-1b"]

   # STS Role ARNs (replace with your actual ARNs)
   installer_role_arn = "arn:aws:iam::123456789012:role/my-HCP-ROSA-Installer-Role"
   support_role_arn = "arn:aws:iam::123456789012:role/my-HCP-ROSA-Support-Role"
   worker_role_arn = "arn:aws:iam::123456789012:role/my-HCP-ROSA-Worker-Role"
   operator_role_prefix = "my-operator"
   oidc_config_id = "my-oidc-config-id"

   # Optional: Customize image mirrors
   image_mirrors = {
     "docker.io/library/nginx" = ["registry.corp.example.com/nginx"]
     "docker.io/library/redis" = ["registry.corp.example.com/redis", "quay.io/backup/redis"]
   }

   # Optional: Add cluster tags
   cluster_tags = {
     Environment = "production"
     Team = "platform"
   }
   ```

2. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Variables

### Required Variables

| Name | Description | Type |
|------|-------------|------|
| cluster_name | Name of the ROSA HCP cluster | string |
| aws_account_id | AWS account ID | string |
| aws_billing_account_id | AWS billing account ID | string |
| subnet_ids | List of subnet IDs | list(string) |
| availability_zones | List of availability zones | list(string) |
| installer_role_arn | ARN of the installer role | string |
| support_role_arn | ARN of the support role | string |
| worker_role_arn | ARN of the worker role | string |
| operator_role_prefix | Prefix for operator roles | string |
| oidc_config_id | OIDC configuration ID | string |

### Optional Variables

| Name | Description | Type | Default |
|------|-------------|------|---------|
| aws_region | AWS region | string | "us-east-1" |
| replicas | Number of worker nodes | number | 2 |
| openshift_version | OpenShift version | string | "4.15.9" |
| cluster_properties | Additional cluster properties | map(string) | {} |
| cluster_tags | Cluster tags | map(string) | {} |
| image_mirrors | Image mirror configurations | map(list(string)) | See default |

## Outputs

### Cluster Information
- `cluster_id`: Unique cluster identifier
- `cluster_name`: Cluster name
- `cluster_state`: Current cluster state
- `cluster_api_url`: Kubernetes API endpoint
- `cluster_console_url`: OpenShift web console URL
- `cluster_domain`: Cluster domain

### Image Mirror Information
- `configured_image_mirrors`: Details of all image mirrors
- `image_mirror_count`: Total number of mirrors
- `mirrored_sources`: List of mirrored source registries
- `all_cluster_mirrors`: All mirrors (including pre-existing)

## Deployment Flow

1. **Cluster Creation**: ROSA HCP cluster is created with STS roles
2. **Wait for Readiness**: Terraform waits for cluster to be fully ready
3. **Image Mirror Configuration**: Mirrors are configured after cluster is ready
4. **Verification**: Data source retrieves all mirrors for verification

## Corporate Registry Example

For a corporate environment, you might configure:

```hcl
image_mirrors = {
  # Mirror public Docker Hub images through corporate registry
  "docker.io/library/nginx" = ["registry.corp.example.com/docker/nginx"]
  "docker.io/library/postgres" = ["registry.corp.example.com/docker/postgres"]
  "docker.io/library/redis" = ["registry.corp.example.com/docker/redis"]

  # Mirror Red Hat images
  "registry.redhat.io/ubi8/ubi" = ["registry.corp.example.com/redhat/ubi8"]

  # Mirror Quay.io images
  "quay.io/prometheus/prometheus" = ["registry.corp.example.com/quay/prometheus"]
  "quay.io/grafana/grafana" = ["registry.corp.example.com/quay/grafana"]
}
```

## High Availability Configuration

For production environments, configure multiple mirrors:

```hcl
image_mirrors = {
  "docker.io/library/nginx" = [
    "registry-primary.corp.example.com/nginx",    # Primary
    "registry-backup.corp.example.com/nginx",     # Backup
    "quay.io/emergency/nginx"                     # Emergency fallback
  ]
}
```

## Cleanup

To destroy the resources:

```bash
terraform destroy
```

**Note**: The cluster destruction will automatically remove all associated image mirrors.

## Expected Timeline

- **Cluster Creation**: 15-20 minutes
- **Image Mirror Configuration**: 1-2 minutes per mirror
- **Total Deployment Time**: ~20-25 minutes

## Troubleshooting

1. **Cluster Creation Fails**: Check STS role permissions and OIDC configuration
2. **Image Mirror Fails**: Verify cluster is HCP type and fully ready
3. **Network Issues**: Ensure subnets have proper routing and security groups

## Next Steps

After deployment:

1. Configure additional OpenShift resources (monitoring, logging, etc.)
2. Set up application deployments that will use the mirrored images
3. Test image pulling from configured mirrors
4. Monitor cluster and mirror health