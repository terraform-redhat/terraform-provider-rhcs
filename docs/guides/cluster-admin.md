---
page_title: "Creating Cluster Admin ROSA Cluster"
subcategory: ""
description: |-
  Instructions on how to create a Rosa Openshift cluster admin when you create your cluster with the terraform provider.
---

# Creating a cluster admin user for your ROSA cluster

You can create a Cluster Admin user for your cluster using Terraform.

## Prerequisites

1. You created your [account roles using Terraform](../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).

## Creating a cluster administrator by using Terraform

You must implement the following parameters to create your Cluster Admin role on cluster creation using the htpsswd identity provider.

| Parameter | Type | Description |
|-----------|------|-------------|
| `admin_credentials` | Attributes | This attribute list defines the user name and password for a cluster admin user. |
| `password` | Case-sensitive String | Enter the cluster admin's password that will be created with the cluster. NOTE: This field is case sensitive. |
| `username` | String | Enter your cluster admin's username that will be created with the cluster. |

1. Add these parameters to your cluster creation `variables.tf` and `main.tf` files.
1. Run `terraform apply` to upgrade your cluster.

## OpenShift documentation

 - [Configuring an htpasswd identity provider](https://docs.openshift.com/rosa/rosa_install_access_delete_clusters/rosa-sts-config-identity-providers.html#config-htpasswd-idp_rosa-sts-config-identity-providers)
