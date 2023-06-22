# Creating and modifying machine pools on your cluster

You can do the following cluster administrative tasks.

## Prerequisites

1. You created your [account roles using Terraform](../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).

## Common machine pool tasks

You may need to perform any of the following administrative tasks to create or modify non-default machine pools:

### Create a new machine pool

If you need to create a machine pool, use the `create_machine_pool` [Terraform module](../examples/create_machine_pool/README.md).

### Changing replica count

The replica count specifies how many compute nodes you want to provision. Using the `create_machine_pool` [Terraform module](../examples/create_machine_pool/README.md), you can change the replica count with the `replica` variable. The default replica count is **2**.
```
export TF_VAR_replica=<integer>
```

### Auto Scaling 

You may enable or disable autoscaling on your machine pools. You must set a minimum or maximum replica count variables for Terraform. The autoscaling will not exceed whichever value you set. For more information, see [About autoscaling nodes on a cluster](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/cluster_administration/nodes#rosa-nodes-about-autoscaling-nodes) in the Red Hat Customer Portal.

### Changing labels on machine pools

You may add or edit labels on compute nodes. For more information, see [Adding node labels to a machine pool](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/cluster_administration/nodes#rosa-adding-node-labels_rosa-managing-worker-nodes) in the Red Hat Customer Portal.