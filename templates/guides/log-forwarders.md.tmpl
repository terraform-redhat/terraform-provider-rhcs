---
page_title: "Configure Log Forwarders for ROSA HCP Clusters"
subcategory: ""
description: |-
  Instructions on how to configure log forwarders for ROSA HCP clusters as standalone resources and at cluster creation.
---

# Configuring Log Forwarders

Log forwarders enable you to forward cluster logs to external destinations such as S3 or CloudWatch. You can manage log forwarders as standalone resources after cluster creation or configure them at cluster creation time.

## Prerequisites

1. You have created your ROSA HCP cluster using Terraform or are planning to create one.
2. You have configured the necessary AWS resources (S3 bucket or CloudWatch log group) for log forwarding. For instructions, see [Creating an IAM role and policy](https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/4/html/security_and_compliance/rosa-forwarding-control-plane-logs#rosa-create-an-iam-role-policy_rosa-configuring-the-log-forwarder).
3. You have the required IAM permissions to access the log destination.

## Managing Log Forwarders After Cluster Creation

You can create, update, and delete log forwarders as standalone resources after cluster creation.

### Creating a Log Forwarder

When configuring log forwarders, you can specify which applications and groups to forward logs from. For a complete list of available applications and groups, see [Determining what log groups to use](https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/4/html/security_and_compliance/rosa-forwarding-control-plane-logs#rosa-determine-log-groups_rosa-configuring-the-log-forwarder).

#### S3 Destination

```terraform
resource "rhcs_log_forwarder" "s3_forwarder" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id

  s3 = {
    bucket_name   = "my-logs-bucket"
    bucket_prefix = "rosa-logs/"
  }

  applications = ["application-1", "application-2"]
}
```

#### CloudWatch Destination

```terraform
resource "rhcs_log_forwarder" "cloudwatch_forwarder" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id

  cloudwatch = {
    log_group_name            = "/aws/rosa/my-cluster"
    log_distribution_role_arn = "arn:aws:iam::123456789012:role/my-log-forwarder-role"
  }

  applications = ["application-1"]
}
```

#### Using Log Forwarder Groups

Log forwarder groups allow you to forward logs for a group of applications without needing to specify each individual application in that group. The `version` field is optional; if not specified, the backend will use the latest version.

For a complete list of available groups and applications, see [Determining what log groups to use](https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/4/html/security_and_compliance/rosa-forwarding-control-plane-logs#rosa-determine-log-groups_rosa-configuring-the-log-forwarder).

```terraform
resource "rhcs_log_forwarder" "grouped_forwarder" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id

  s3 = {
    bucket_name = "my-logs-bucket"
  }

  groups = [
    {
      id = "group-name"
    },
  ]
}
```

### Updating a Log Forwarder

To update a log forwarder, modify the resource configuration and run `terraform apply`:

```terraform
resource "rhcs_log_forwarder" "s3_forwarder" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id

  s3 = {
    bucket_name   = "my-logs-bucket"
    bucket_prefix = "updated-prefix/"
  }

  # Update applications list
  applications = ["application", "new-application"]
}
```

### Deleting a Log Forwarder

To delete a specific log forwarder while keeping your cluster and other resources, remove the `rhcs_log_forwarder` resource from your Terraform configuration and run `terraform apply`. Terraform will detect the removal and delete only that log forwarder.

## Configuring Log Forwarders at Cluster Creation

You can configure log forwarders during cluster creation to forward cluster installation logs by using the `log_forwarders_at_cluster_creation` attribute. After cluster creation, these log forwarders can be imported as standalone log forwarders so that they can be edited or deleted.

### Step 1: Create Cluster with Log Forwarders

Add the `log_forwarders_at_cluster_creation` block to your cluster resource configuration:

```terraform
resource "rhcs_cluster_rosa_hcp" "rosa_hcp_cluster" {
  name                         = "my-cluster"
  version                      = "4.20.8"
  channel_group                = "stable"
  cloud_region                 = "us-west-2"
  aws_account_id               = "123456789012"
  aws_billing_account_id       = "123456789012"
  availability_zones           = ["us-west-2a"]
  replicas                     = 3
  compute_machine_type         = "m5.xlarge"
  aws_subnet_ids               = ["subnet-1a2b3c4d5e6f7g8h9", "subnet-9h8g7f6e5d4c3b2a1"]
  wait_for_create_complete     = true
  
  sts = {
    operator_role_prefix = "my-cluster-operator"
    role_arn             = "arn:aws:iam::123456789012:role/HCP-ROSA-Installer-Role"
    support_role_arn     = "arn:aws:iam::123456789012:role/HCP-ROSA-Support-Role"
    instance_iam_roles = {
      worker_role_arn = "arn:aws:iam::123456789012:role/HCP-ROSA-Worker-Role"
    }
    oidc_config_id = "2abcdef34567890abcdef1234567890a"
  }

  properties = {
    rosa_creator_arn = "arn:aws:iam::123456789012:user/my-user"
  }

  log_forwarders_at_cluster_creation = [
    {
      s3 = {
        bucket_name = "my-logs-bucket"
      }
      applications = ["kube-apiserver"]
      groups = [
        {
          id = "group-name"
        }
      ]
    }
  ]
}
```

Run `terraform apply` to create the cluster with the log forwarder configuration.

### Step 2: View Log Forwarder IDs

After the cluster enters the ready state, run `terraform show` to view the `log_forwarder_ids` attribute. This attribute contains a list of all log forwarder IDs associated with the cluster:

```bash
terraform show
```

You should see output similar to:

```
log_forwarder_ids = [
  "2abcdef34567890abcdef1234567890c",
]
```

### Step 3: Import Log Forwarder for Standalone Management

To manage the log forwarder as a standalone resource (see section above), import it using the cluster ID and log forwarder ID:

```bash
terraform import rhcs_log_forwarder.my_log_forwarder <cluster_id>,<log_forwarder_id>
```

Example:

```bash
terraform import rhcs_log_forwarder.my_log_forwarder 2abcdef34567890abcdef1234567890b,2abcdef34567890abcdef1234567890c
```

### Step 4: Manage Log Forwarder as Standalone Resource

After importing, create a `rhcs_log_forwarder` resource in your Terraform configuration. The configuration should match the initial cluster creation settings:

```terraform
resource "rhcs_log_forwarder" "my_log_forwarder" {
  cluster = rhcs_cluster_rosa_hcp.rosa_hcp_cluster.id

  s3 = {
    bucket_name   = "my-logs-bucket"
    bucket_prefix = ""
  }

  applications = ["kube-apiserver"]
  groups = [
    {
      id = "group-name"
    }
  ]
}
```

You can now manage the log forwarder using standard Terraform operations as described in the section above:

- **Update**: Modify the configuration and run `terraform apply`
- **Delete**: Remove the resource block and run `terraform apply`

## Configuration Reference

### log_forwarders_at_cluster_creation

- **Type**: List of objects
- **Optional**: Yes
- **Immutable**: Yes (cannot be modified after cluster creation)

Each log forwarder object supports:

- `s3` - (Optional) S3 configuration for log forwarding destination
  - `bucket_name` - (Required) The name of the S3 bucket (e.g., `my-logs-bucket`)
  - `bucket_prefix` - (Optional) The prefix to use for objects stored in the S3 bucket (e.g., `rosa-logs/`)

- `cloudwatch` - (Optional) CloudWatch configuration for log forwarding destination
  - `log_group_name` - (Required) The name of the CloudWatch log group (e.g., `/aws/rosa/my-cluster`)
  - `log_distribution_role_arn` - (Required) The ARN of the IAM role for log distribution (e.g., `arn:aws:iam::123456789012:role/my-log-forwarder-role`)

- `applications` - (Optional) List of additional applications to forward logs for (e.g., `["audit", "infrastructure"]`)

- `groups` - (Optional) List of log forwarder groups
  - `id` - (Required) The identifier of the log forwarder group (e.g., `api`, `scheduler`, `authentication`)
  - `version` - (Optional) The version of the log forwarder group (e.g., `1.0`, `2.0`). If not specified, the backend will use the latest version.

**Note**: Either `s3` or `cloudwatch` must be specified, but not both.

### rhcs_log_forwarder Resource

For detailed information on the `rhcs_log_forwarder` resource and its attributes, see the [resource documentation](https://registry.terraform.io/providers/terraform-redhat/rhcs/latest/docs/resources/log_forwarder).

## Additional Notes

- The `log_forwarders_at_cluster_creation` field will not appear in `terraform show` output after cluster creation.
- Use the `log_forwarder_ids` computed attribute to view all log forwarders associated with your cluster.
- Log forwarders configured at cluster creation must be imported before they can be managed as standalone resources.
- You cannot configure both S3 and CloudWatch destinations in the same log forwarder.

## OpenShift Documentation

- [ROSA Logging](https://docs.openshift.com/rosa/observability/logging/rosa-logging.html)
- [Forwarding logs to external systems](https://docs.openshift.com/rosa/observability/logging/log-forwarding.html)
