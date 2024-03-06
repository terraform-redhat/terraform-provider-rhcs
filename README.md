---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "Red Hat Cloud Services Terraform Provider"
subcategory: ""
description: |-
  
---
<a href="https://redhat.com">
    <img src=".github/Logo_Red_Hat.png" alt="Red Hat logo" title="Red Hat" align="right" max-width="60px" />
</a>

# Red Hat Cloud Services Terraform Provider

> Please note that this Terraform provider and its modules are open source and will continue to iterate features, gradually maturing this code.
> If you encounter any issues, please report them in this repo.

## Introduction

The Red Hat Cloud Services Terraform provider allows Terraform to manage Red Hat OpenShift Service on AWS (ROSA) clusters and relevant resources.

For more information about ROSA, see the Red Hat documentation [here](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/introduction_to_rosa/rosa-understanding).

## Prerequisites 
* [GoLang version 1.20 or newer](https://go.dev/doc/install)
* [Terraform version 1.4.6 or newer](https://developer.hashicorp.com/terraform/downloads)
* An offline [OCM token](https://console.redhat.com/openshift/token/rosa)
* [AWS account](https://aws.amazon.com/console/)
* Completed [the ROSA getting started](https://console.redhat.com/openshift/create/rosa/getstarted) requirements
* [ROSA CLI](https://console.redhat.com/openshift/downloads#tool-rosa)
* **Optional**: A [configured `*.tfvars` file](docs/guides/terraform-vars.md).

## Provider documentation

See [the Terraform Registry documentation](https://registry.terraform.io/providers/terraform-redhat/rhcs/latest/docs) for instructions on using this provider.

## Limitations of the OCM Terraform provider

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
* [Account Roles Terraform](examples/create_account_roles/)
### Cluster creation examples
* [Create a ROSA cluster that uses STS and has a managed OIDC configuration](examples/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/)
* [Create a ROSA cluster that uses STS and has an unmanaged OIDC configuration](examples/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/)

### Post cluster installation
* Modifying default machine pools
* The following [identity providers](examples/create_identity_provider/) are supported:
  * [Github](examples/create_identity_provider/github/)
  * [Gitlab](examples/create_identity_provider/gitlab/)
  * [HTPasswd](examples/create_identity_provider/htpasswd/)
  * [Google](examples/create_identity_provider/google/)
  * [LDAP](examples/create_identity_provider/ldap/)
* [Upgrading or updating your cluster](docs/guides/upgrading-cluster.md)

## Contributing to the Red Hat Cloud Service Terraform provider
If you want to build a local Red Hat Cloud Services provider to develop improvements for the Red Hat Cloud Services provider, you can run `terraform plan` against your local build with:
1. Run  ```make install``` in the repository root directory. After running ```make install``` you will find the rhcs provider binary file in the directory:
    ```
    <HOME>/.terraform.d/plugins/terraform.local/local/rhcs/<VERSION>/<TARGET_ARCH>
    ```

    For example, the following location would contain the `terraform-rhcs-provider` binary file: 
    ```    
    ~/.terraform.d/plugins/terraform.local/local/rhcs/0.0.1/linux_amd64
    ```
  
2. You now need to update your `main.tf` to the location of the local provider by pointing the required_providers rhcs to the local terraform directory.

    ```
    terraform {
      required_providers {
        rhcs = {
          source  = "terraform.local/local/rhcs"
          version = ">=0.0.1"
        }
      }
    }

    provider "rhcs" {
      token = var.token
      url = var.url
    }
    ```

### Testing binary
If you want to locally test the provider binary without building from sources, you can pull the `latest` binary container image and copy the binary file to your local by running `make binary` or the whole background command if you need to make custom changes:
```
    podman run --pull=always --rm registry.ci.openshift.org/ci/rhcs-tf-bin:latest cat /root/terraform-provider-rhcs > ~/terraform-provider-rhcs && chmod +x ~/terraform-provider-rhcs
```
You can also use specific commit images by substituting `latest` for the desired commit SHA.
Binary image only runs on AMD64 architectures up to now.

### Developing the Provider
Detailed documentation for developing and contributing to RHCS provider can be found in our [contribution guide](CONTRIBUTE.md).
