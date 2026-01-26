---
page_title: "Trust Policy External ID for ROSA Clusters"
subcategory: ""
description: |-
  Guide for configuring trust policy external ID in ROSA clusters using the RHCS Terraform provider.
---

# Trust Policy External ID

The `trust_policy_external_id` field provides an external ID for trust policy condition in account roles for ROSA clusters. This field is available in both ROSA Classic and ROSA HCP cluster configurations.

## Configuration

The `trust_policy_external_id` is configured within the `sts` block of your ROSA cluster resource:

### ROSA HCP Example

```terraform
resource "rhcs_cluster_rosa_hcp" "example" {
  name           = "my-cluster"
  cloud_region   = "us-east-2"
  aws_account_id = "123456789012"
  # ... other required fields

  sts = {
    role_arn                  = "arn:aws:iam::123456789012:role/my-installer-role"
    support_role_arn         = "arn:aws:iam::123456789012:role/my-support-role"
    operator_role_prefix     = "my-operator-prefix"
    trust_policy_external_id = "unique-external-id-123"
    
    instance_iam_roles = {
      worker_role_arn = "arn:aws:iam::123456789012:role/my-worker-role"
    }
  }
}
```

### ROSA Classic Example

```terraform
resource "rhcs_cluster_rosa_classic" "example" {
  name           = "my-cluster"
  cloud_region   = "us-east-2"
  aws_account_id = "123456789012"
  # ... other required fields

  sts = {
    role_arn                  = "arn:aws:iam::123456789012:role/my-installer-role"
    support_role_arn         = "arn:aws:iam::123456789012:role/my-support-role"
    operator_role_prefix     = "my-operator-prefix"
    trust_policy_external_id = "unique-external-id-123"
    
    instance_iam_roles = {
      master_role_arn = "arn:aws:iam::123456789012:role/my-master-role"
      worker_role_arn = "arn:aws:iam::123456789012:role/my-worker-role"
    }
  }
}
```

## Field Properties

- **Type**: String
- **Required**: No (Optional)
- **Description**: External ID for trust policy condition in account roles

## Useful Resources

- [ROSA Classic Cluster Resource](../resources/cluster_rosa_classic.md)
- [ROSA HCP Cluster Resource](../resources/cluster_rosa_hcp.md)
- [AWS IAM External ID Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html)
