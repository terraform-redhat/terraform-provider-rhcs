---
page_title: "Upgrade Rosa Openshift Cluster"
subcategory: ""
description: |-
  Instructions on how to upgrade Rosa Openshift cluster created by the terraform provider.
---

# Updating or upgrading your ROSA cluster

You can update or upgrade your cluster using Terraform.

## Prerequisites

1. You created your [account roles using Terraform](../examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md).
1. You created your cluster using Terraform. This cluster can either have [a managed OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md) or [an unmanaged OIDC configuration](../examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md).

## Upgrading your cluster

To upgrade your ROSA cluster to another version, upgrade your account roles and policies, export the following variables, and then run `terraform apply`.

1. If you are upgrading to a new y-stream, such as upgrading from version 4.12.19 to 4.13.1, run `terraform apply` in the [`account_roles` module](../../examples/create_account_roles/README.md). You must specify your desired y-stream version by using the following variables: 

1. Export the `TF_VAR_openshift_version` with the intended version. Your value must be prepended with `openshift-v`.
        ```
        export TF_VAR_openshift_version=<version_number>
        ```
1. Upgrading your cluster requires approval, especially when transitioning between major y-streams. You may be required to provide administrative confirmation regarding significant modifications to your cluster. In this case, when you first attempt the upgrade, you will receive an error message that provides guidance regarding the necessary modifications. It is essential to follow the instructions carefully. Indicate completion of the requirements by adding the `upgrade_acknowledgements_for` variable to your Terraform plan with your targeted version. For example, if you are upgrading from version 4.11.43 to 4.12.21, you should use '4.12' as the value for this variable.
        ```
        upgrade_acknowledgements_for = <version_acknowledgement>
        ```
1. Run `terraform apply` to upgrade your cluster.

## OpenShift documentation

 - [Upgrading ROSA clusters with STS](hhttps://docs.openshift.com/rosa/upgrading/rosa-upgrading-sts.html)
