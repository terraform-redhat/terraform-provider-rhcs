---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ocm Provider"
subcategory: ""
description: |-
  
---
<a href="https://redhat.com">
    <img src=".github/Logo_Red_Hat.png" alt="RedHat logo" title="RedHat" align="right" height="50" />
</a>

# Red Hat Cloud Services Terraform Provider

> Please note that this Terraform provider and its modules are open source and will continue to iterate features, gradually maturing this code.
> If you encounter any issues, please report them in this repo.

## Introduction

The Red Hat OCM provider allows Terraform to manage Red Hat OpenShift Service on AWS (ROSA) clusters, machine pools, and an identity provider.

For more information about ROSA, see the Red Hat documentation [here](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/introduction_to_rosa/rosa-understanding).

## Prerequisites 
* [GoLang version 1.20 or newer](https://go.dev/doc/install)
* [Terraform version 1.4.6 or newer](https://developer.hashicorp.com/terraform/downloads)
* An offline [OCM token](https://console.redhat.com/openshift/token/rosa)
* [AWS account](https://aws.amazon.com/console/)
* Completed [the ROSA getting started](https://console.redhat.com/openshift/create/rosa/getstarted) requirements
* [ROSA CLI](https://console.redhat.com/openshift/downloads#tool-rosa)
* **Optional**: A [configured `*.tfvars` file](docs/terraform-vars.md).

## Provider Documentation

See [the Terraform Registry documentation](https://registry.terraform.io/providers/terraform-redhat/ocm/latest/docs) for instructions on using this provider.

## Limitations of OCM Terraform provider

The following items are limitations with the current release of the OCM Terraform provider:

* The latest version is not backward compatible with version 1.0.1.
* When creating a cluster, the cluster uses AWS credentials configured on your local machine. These credentials provide access to the AWS API for validating your account.
* When creating a machine pool, you need to specify your replica count. You must define either the `replicas= "<count>"` variable or provide values for the following variables to build the machine pool:  
   * `min_replicas = "<count>"` 
   * `max_replicas="<count>"` 
   * `autoscaling_enabled=true`
* The htpasswd identity provider does not support creating the identity provider with multiple users or adding additional users to the existing identity provider.
* The S3 bucket that is created as part of the OIDC configuration must be created in the same region as your OIDC provider.
* The Terraform provider does not support auto-generated `operator_role_prefix`. You must provide your `operator_role_prefix` when creating the account roles.

## Examples

The example Terraform files are all considered in development:
### Prior to creating clusters
1. [Account Roles terraform](https://github.com/terraform-redhat/terraform-provider-ocm/blob/a42779d6b6712f4dde358344f44b782e4dfcd120/examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/README.md)
### Cluster Creation examples
1. [Create a ROSA cluster that usess STS and has a managed OIDC configuration](https://github.com/terraform-redhat/terraform-provider-ocm/blob/529b10c22c13810b3edbaa01327c7dfdcc207650/examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/README.md)
1. [Create a ROSA cluster that uses STS and has an unmanaged OIDC configuration](https://github.com/terraform-redhat/terraform-provider-ocm/blob/529b10c22c13810b3edbaa01327c7dfdcc207650/examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/README.md)

### Post cluster installation
1. Modifying default machine pools
1. [Identity provider](/examples/create_identity_provider/README.md). The following identity providers are supported:
      * [Github](/examples/create_identity_provider/github/README.md)
      * [Gitlab](/examples/create_identity_provider/gitlab/README.md)
      * [HTPasswd](/examples/create_identity_provider/htpasswd/README.md)
      * [Google](/examples/create_identity_provider/google/README.md)
      * [LDAP](/examples/create_identity_provider/ldap/README.md)
1. [Upgrading or Updating your cluster](docs/upgrading-cluster.md)

## Contributing to the Red Hat OCM Terraform Provider
If you want to build a local OCM provider to develop improvements for the Red Hat OCM provider, you can run `terraform plan` against your local build with:
1. Run  ```make install``` in the repository root directory. After running ```make install``` you will find the ocm provider binary file in the directory:
    ```
    <HOME>/.terraform.d/plugins/terraform.local/local/ocm/<VERSION>/<TARGET_ARCH>
    ```

    For example, the following location would contain the `terraform-ocm-provider` binary file: 
    ```    
    ~/.terraform.d/plugins/terraform.local/local/ocm/0.0.1/linux_amd64
2. You now need to update your `main.tf` to the location of the local provider  by pointing the required_providers ocm to the local terraform directory.

    ```
    terraform {
      required_providers {
        ocm = {
          source  = "terraform.local/local/ocm"
          version = ">=0.0.1"
        }
      }
    }

    provider "ocm" {
      token = var.token
      url = var.url
    }

### Testing binary
If you want to locally test the provider binary without building from sources, you can pull the `latest` container image and copy the binary from the directory :
    
    <HOME>/.terraform.d/plugins/terraform.local/local/ocm/<VERSION>/<TARGET_ARCH>
    
to your local using the following example: 

    podman run --pull=always --rm registry.ci.openshift.org/ci/ocm-tf-e2e:latest cat /root/.terraform.d/plugins/terraform.local/local/ocm/1.0.1/linux_amd64/terraform-provider-ocm > ~/terraform-provider-ocm && chmod +x ~/terraform-provider-ocm
    can also use specific commit images by substituting `latest` for the desired commit SHA.
