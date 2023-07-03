---
page_title: "Create and modify Machine Pool"
subcategory: ""
description: |-
  Instructions on how to create and modify Machine Pools.
---

# Using machine pools on your cluster

## Prerequisites

1. You created your [account roles using Terraform](../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).

## Common machine pool tasks

You may need to perform any of the following administrative tasks to create or modify non-default machine pools:

### Create a new machine pool

If you need to create a machine pool, use the `create_machine_pool` [Terraform module](../examples/create_machine_pool/README.md).

1. Export the number of replicas for your machine pool:
    ```
    export TF_VAR_replica=<integer>
    ```
1. Export the ID of the cluster where you want to create a machine pool. This ID can be found with the `rosa list cluster` command:
    ```
    export TF_VAR_cluster_id=<cluster_ID_string>
    ```
1. Export the name of your new machine pool:
    ```
    export TF_VAR_name=<name_of_machine_pool>
    ```
1. Export the machine type of your machine pool:
    ```
    export TF_VAR_machine_type=<machine_type>
    ```
1. **Optional**: Export a label by using a key-value pair with commas separating each label
    ````
    export TF_VAR_labels='{ test="first-label" }'
1. Run `terraform apply` to update the machine pools:
    ```
    terraform apply
    ```
1. Your machine pools should be updated.
### Changing replica count

The replica count specifies how many compute nodes you want to provision. Using the `create_machine_pool` [Terraform module](../examples/create_machine_pool/README.md), you can change the replica count with the `replica` variable.

To change the replica count within your default machine pool, run these commands in the same Terraform module where you created your cluster. See the [Cluster with Managed OIDC configurations](examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) and [Cluster with Unmanaged OIDC configurations](examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md) for examples. To set the replica count on an additional machine pool, export the machine pool name with `export TF_VAR_name=<name_of_machine_pool> `first. See [Creating a machine pool for an existing OpenShift cluster](examples/create_machine_pool/README.md) for examples.

1. Export the new replica count:
    ```
    export TF_VAR_replica=<integer>
    ```
1. Run `terraform apply` to update the machine pools.
    ```
    terraform apply
    ```
1. Your machine pools should be updated.
### Autoscaling 

You may enable or disable autoscaling on your machine pools. You must set either the minimum or maximum replica count variable for Terraform. The autoscaling will not exceed whichever value you set. For more information, see [About autoscaling nodes on a cluster](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/cluster_administration/nodes#rosa-nodes-about-autoscaling-nodes) in the Red Hat Customer Portal.

To enable autoscaling within your default machine pool, run these commands in the same Terraform module where you created your cluster. See the [Cluster with Managed OIDC configurations](examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) and [Cluster with Unmanaged OIDC configurations](examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md) for examples. To enable autoscaling on an additional machine pool, export the machine pool name with `export TF_VAR_name=<name_of_machine_pool> `first. See [Creating a machine pool for an existing OpenShift cluster](examples/create_machine_pool/README.md) for examples.
1. Unset the environmental variable for `replicas` to be able to enable autoscaling:
    ```
    unset TF_VAR_replicas

1. Export the following variables for a maximum and minimum replica count as well as enabling autoscaling:
    > **IMPORTANT**: You must specify a maximum and minimum replica count when enabling autoscaling.
    ````
    export TF_VAR_autoscaling_enabled=<true|false>
    ````
    ````
    export TF_VAR_min_replicas=<integer>
    ````
    ````
    export TF_VAR_max_replicas=<integer>
2. Run `terraform apply` to update the machine pools:
    ```
    terraform apply
    ```
3. Your machine pools should be updated to enable or disable autoscaling.
### Changing labels on machine pools

You may add or edit labels on compute nodes. For more information, see [Adding node labels to a machine pool](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/cluster_administration/nodes#rosa-adding-node-labels_rosa-managing-worker-nodes) in the Red Hat Customer Portal.

To add labels to your default machine pool, run these commands in the same Terraform module where you created your cluster. See the [Cluster with Managed OIDC configurations](examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) and [Cluster with Unmanaged OIDC configurations](examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md) for examples. To add labels to an additional machine pool, export the machine pool name with `export TF_VAR_name=<name_of_machine_pool> `first. See [Creating a machine pool for an existing OpenShift cluster](examples/create_machine_pool/README.md) for examples.

1. Export your label by using a key-value pair with commas separating each label:
    ````
    export TF_VAR_labels='{ test="first-label" }'
2. Run `terraform apply` to update the machine pools:
    ```
    terraform apply
    ```
3. Your machine pools should be updated with your provided label.
### Deleting a machine pool
You can use the `terraform destroy` command to delete additional machine pools. You should specify the desired machine pool with the `name` variable.
