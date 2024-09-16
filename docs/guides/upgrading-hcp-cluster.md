---
page_title: "Upgrade ROSA HCP Openshift Cluster or Machine Pool"
subcategory: ""
description: |-
  Instructions on how to upgrade ROSA HCP Openshift cluster or machine pool created via the terraform provider.
---

# Updating or upgrading your ROSA HCP cluster

You can update or upgrade your cluster using Terraform.

## Prerequisites

1. You created your [account roles using Terraform](https://registry.terraform.io/modules/terraform-redhat/rosa-hcp/rhcs/latest/submodules/account-iam-resources).
2. You created your cluster using Terraform.

## Upgrading your cluster

On Hosted Control Plane topology, the upgrade of control plane occurs separately from machine pools, it is important to notice that the machine pool version cannot be greater than that of the control plane.
It is not possible to upgrade both sides to the same version alongside each other as of now.

The following steps applies to both a cluster resource and the machine pool resources.

To upgrade your ROSA cluster to another version, export the following variables, and then run `terraform apply`.

1. Export the `TF_VAR_openshift_version` with the intended version.
        ```
        export TF_VAR_openshift_version=<version_number>
        ```
2. Upgrading your resource may require approval, especially when transitioning between major y-streams. You may be required to provide administrative confirmation regarding significant modifications to your resource. In this case, when you first attempt the upgrade, you will receive an error message that provides guidance regarding the necessary modifications. It is essential to follow the instructions carefully. Indicate completion of the requirements by adding the `upgrade_acknowledgements_for` variable to your Terraform plan with your targeted version. For example, if you are upgrading from version 4.11.43 to 4.12.21, you should use '4.12' as the value for this variable.
        ```
        upgrade_acknowledgements_for = <version_acknowledgement>
        ```
3. Run `terraform apply` to upgrade your cluster or machine pool.

## OpenShift documentation

 - [Upgrading ROSA HCP clusters](https://docs.openshift.com/rosa/upgrading/rosa-hcp-upgrading.html)
