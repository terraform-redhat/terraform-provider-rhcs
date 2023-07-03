---
page_title: "Upgrade Rosa Openshift Cluster"
subcategory: ""
description: |-
  Instructions on how to upgrade Rosa Openshift cluster created by the terraform provider.
---

# Updating or upgrading your ROSA cluster

You can update or upgrade your cluster using Terraform.

## Prerequisites

1. You created your [account roles using Terraform](../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md***REMOVED***.
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md***REMOVED*** or [an unmanaged OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md***REMOVED***.

## Upgrading your cluster

To upgrade your ROSA cluster to another version, export the following variable then run `terraform apply`.

1. Export the `TF_VAR_openshift_version` with the intended version. Your value must be prepended with `openshift-v`.
    ```
    export TF_VAR_openshift_version=<version_number>
    ```
1. Upgrading to a new cluster version, especially when transitioning between major y-streams, requires approval. You might be requested to provide administrative confirmation regarding significant modifications for your cluster. In this case, when you first attempt the upgrade, you will receive an error message that provides guidance on the necessary modifications. It is essential to follow those instructions carefully and indicate completion of the requirements by adding the "upgrade_acknowledgements_for" attribute to your resource, specifying the target version. For example, if you are upgrading from 4.11.43 to 4.12.21, you should use '4.12' as the value for this variable.
    ```
    upgrade_acknowledgements_for = <version_acknowledgement>
    ```
1. Run `terraform apply` to upgrade your cluster.

## OpenShift documentation

 - [Upgrading ROSA clusters with STS](hhttps://docs.openshift.com/rosa/upgrading/rosa-upgrading-sts.html***REMOVED***
