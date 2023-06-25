# Updating or Upgrading your ROSA cluster

You can update or upgrade your cluster using Terraform.

## Prerequisites

1. You created your [account roles using Terraform](../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md***REMOVED***.
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md***REMOVED*** or [an unmanaged OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md***REMOVED***.

## Upgrading your cluster

To upgrade your ROSA cluster to another version, export the following variable then run `terraform apply`.

1. Export the `TF_VAR_openshift_version` with the intended version. Your value must be prepended with `openshift-v` to succeed.
    ```
    export TF_VAR_openshift_version=<version_number>
    ```
1. When upgrading from between major Y-Streams, you may need to provide an administrative acknowledgement that there are major changes for your cluster. For example, if you are upgrading from 4.11.43 to 4.12.21, you must use `4.12` as the value for this variable.
    ```
    export TF_VAR_acks_for = <version_acknowledgement>
    ```
1. Run `terraform apply` to upgrade your cluster.

## OpenShift documentation

 - [Upgrading ROSA clusters with STS](hhttps://docs.openshift.com/rosa/upgrading/rosa-upgrading-sts.html***REMOVED***