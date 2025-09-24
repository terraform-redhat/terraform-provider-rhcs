---
page_title: "Deleting ROSA Clusters with Removed Initial Worker Pools"
subcategory: ""
description: |-
  Instructions on how to interact with `ignore_deletion_error` field of machine pools.
---

# Deleting ROSA Clusters with Removed Initial Worker Pools

When managing ROSA clusters with Terraform, you may encounter situations where the initial worker pools (`workers-[0-n]`) have been deleted, but you still need to delete the cluster. This guide explains how to use the `ignore_deletion_error` field to handle this scenario.

## Overview

The `ignore_deletion_error` field is available on both HCP and Classic machine pool resources. It allows Terraform to remove the machine pool resource from its state while ignoring API errors that occur during the deletion process.

This scenario commonly occurs with HCP clusters when attempting to delete the last remaining machine pool that satisfies the backend requirements. For example, if the initial worker pools have been deleted and replaced with custom pools, trying to delete these custom pools will fail because there are no other pools available to meet the cluster's operational requirements. The API prevents this deletion to maintain cluster functionality, requiring the `ignore_deletion_error` field to proceed with cleanup.

## When to Use ignore_deletion_error

Use this field when all of the following are valid:
- Initial worker pools have been deleted (either through Terraform or outside of Terraform via ROSA CLI, Web UI, or API)
- You need to delete the ROSA cluster through Terraform and the custom pools are impeding the cleanup to proceed (API requires data plane nodes to remain in place during normal operations)

## Configuration

Add the `ignore_deletion_error` field to your machine pool resource:

### HCP Machine Pool
```hcl
resource "rhcs_hcp_machine_pool" "example" {
  # ... other configuration ...
  ignore_deletion_error = true
}
```

### Classic Machine Pool
```hcl
resource "rhcs_machine_pool" "example" {
  # ... other configuration ...
  ignore_deletion_error = true
}
```

## How It Works

When `ignore_deletion_error = true`:

1. Terraform attempts to delete the machine pool via the API
2. If the deletion fails (e.g., pool already deleted), Terraform ignores the error and displays a warning message
3. The resource is removed from Terraform state
4. Subsequent cluster deletion can proceed normally

### Warning Message

When the field is used and a deletion error occurs, Terraform will display a warning message:

```
Warning: Cannot delete machine pool

An error occurred deleting the pool, because ignore deletion error is set it will still be removed from the terraform state. Reason: [error details]
```

## Important Considerations

### Risk of Orphaned Resources
⚠️ **Warning**: If you use `ignore_deletion_error` without subsequently deleting the cluster:
- Machine pools may remain active in the backend
- Manual cleanup will be required via ROSA CLI or Web UI
- This could result in unexpected billing

### Safe Usage Pattern
The recommended workflow is:
1. Set `ignore_deletion_error = true` on affected machine pools
2. Apply the configuration to signal intention to ignore deletion errors
3. **Immediately follow up** with cluster deletion (`terraform destroy`)

### Backend Context
When deleting a cluster, the backend service maintains context of associated pools and handles cleanup of any remaining resources automatically. This is why the follow-up cluster deletion is safe and recommended.

## Example Workflow

1. **Identify affected pools**: Determine which machine pools are encountering deletion errors during `terraform destroy`. These are typically pools that cannot be deleted because they are the last remaining pools satisfying the cluster's backend requirements
2. **Update configuration**: Add `ignore_deletion_error = true` to those machine pool resources
3. **Apply changes**: Run `terraform apply` to set the field signaling intention to ignore deletion errors
4. **Delete resources**: Run `terraform destroy` to clean up the resources managed by the terraform state (this is when pools are actually removed from state)

```bash
# Step 3: Apply configuration changes
terraform apply

# Step 4: Delete the resources
terraform destroy
```

## Alternative Manual Cleanup

If you cannot delete the cluster immediately after using `ignore_deletion_error`, you can manually clean up orphaned pools using:

- **ROSA CLI**: `rosa delete machinepool <pool-name> --cluster <cluster-name>`
- **OpenShift Web Console**: Navigate to cluster → Machine pools → Delete

## See Also

- [Machine Pool Guide](machine-pool.md)
- [Worker Machine Pool Guide](worker-machine-pool.md)