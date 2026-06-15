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

## Create-time validation

On cluster create, the provider reads installer and support role trust policies from AWS and validates `sts.trust_policy_external_id`:

- When set, the value must appear in both role trust policies.
- When omitted, create fails if the roles define an external ID that must be declared explicitly. The error includes the discovered value when unambiguous, for example: `set sts.trust_policy_external_id = "your-external-id"`.
- When omitted and the roles define conflicting external IDs, create fails with an error asking you to set the attribute explicitly.
- When omitted and the roles define ambiguous external IDs, create fails with an error asking you to set the attribute explicitly.

Use the same value on the account IAM module and the cluster resource so Terraform configuration, IAM, and OCM stay aligned.

## Useful Resources

- [ROSA Classic Cluster Resource](../resources/cluster_rosa_classic.md)
- [ROSA HCP Cluster Resource](../resources/cluster_rosa_hcp.md)
- [AWS IAM External ID Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html)

## Account roles modules

When creating account roles with Terraform modules, pass the same value to `trust_policy_external_id` on the account IAM module and to `sts.trust_policy_external_id` on the cluster resource. Use the module for your cluster type:

- ROSA Classic: `terraform-redhat/rosa-classic/rhcs//modules/account-iam-resources`
- ROSA HCP: `terraform-redhat/rosa-hcp/rhcs//modules/account-iam-resources`

Example for ROSA HCP:

```terraform
module "create_account_roles" {
  source  = "terraform-redhat/rosa-hcp/rhcs//modules/account-iam-resources"
  version = ">=1.6.3"

  account_role_prefix      = var.account_role_prefix
  trust_policy_external_id = var.trust_policy_external_id
}

resource "rhcs_cluster_rosa_hcp" "example" {
  # ...
  sts = {
    role_arn                  = module.create_account_roles.account_role_arns.installer
    support_role_arn          = module.create_account_roles.account_role_arns.support
    operator_role_prefix      = module.create_account_roles.account_role_prefix
    trust_policy_external_id  = var.trust_policy_external_id
    instance_iam_roles = {
      worker_role_arn = module.create_account_roles.account_role_arns.worker
    }
  }
}
```

For ROSA Classic, use `rhcs_cluster_rosa_classic` and the `terraform-redhat/rosa-classic/rhcs//modules/account-iam-resources` module with the same `trust_policy_external_id` wiring.
