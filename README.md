---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "Red Hat Cloud Services Provider"
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

The Red Hat OCM provider allows Terraform to manage Red Hat OpenShift Service on AWS (ROSA***REMOVED*** clusters, machine pools, and an identity provider.

For more information about ROSA, see the Red Hat documentation [here](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/introduction_to_rosa/rosa-understanding***REMOVED***.

## Prerequisites 
* [GoLang version 1.20 or newer](https://go.dev/doc/install***REMOVED***
* [Terraform version 1.4.6 or newer](https://developer.hashicorp.com/terraform/downloads***REMOVED***
* An offline [OCM token](https://console.redhat.com/openshift/token/rosa***REMOVED***
* [AWS account](https://aws.amazon.com/console/***REMOVED***
* Completed [the ROSA getting started](https://console.redhat.com/openshift/create/rosa/getstarted***REMOVED*** requirements
* [ROSA CLI](https://console.redhat.com/openshift/downloads#tool-rosa***REMOVED***
* **Optional**: A [configured `*.tfvars` file](docs/terraform-vars.md***REMOVED***.

## Provider Documentation

See [the Terraform Registry documentation](https://registry.terraform.io/providers/terraform-redhat/rhcs/latest/docs***REMOVED*** for instructions on using this provider.

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
1. [Account Roles terraform](/examples/create_rosa_cluster/create_rosa_sts_cluster/classic_sts/account_roles/***REMOVED***
### Cluster Creation examples
1. [Create a ROSA cluster that usess STS and has a managed OIDC configuration](/examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_managed_oidc_config/***REMOVED***
1. [Create a ROSA cluster that uses STS and has an unmanaged OIDC configuration](/examples/create_rosa_cluster/create_rosa_sts_cluster/oidc_configuration/cluster_with_unmanaged_oidc_config/***REMOVED***

### Post cluster installation
1. Modifying default machine pools
1. [Identity provider](/examples/create_identity_provider/***REMOVED***. The following identity providers are supported:
      * [Github](/examples/create_identity_provider/github/***REMOVED***
      * [Gitlab](/examples/create_identity_provider/gitlab/***REMOVED***
      * [HTPasswd](/examples/create_identity_provider/htpasswd/***REMOVED***
      * [Google](/examples/create_identity_provider/google/***REMOVED***
      * [LDAP](/examples/create_identity_provider/ldap/***REMOVED***
1. [Upgrading or Updating your cluster](docs/upgrading-cluster.md***REMOVED***

## Contributing to the Red Hat Cloud Service Terraform Provider
If you want to build a local RHCS provider to develop improvements for the Red Hat RHCS provider, you can run `terraform plan` against your local build with:
1. Run  ```make install``` in the repository root directory. After running ```make install``` you will find the rhcs provider binary file in the directory:
    ```
    <HOME>/.terraform.d/plugins/terraform.local/local/rhcs/<VERSION>/<TARGET_ARCH>
    ```

    For example, the following location would contain the `terraform-rhcs-provider` binary file: 
    ```    
    ~/.terraform.d/plugins/terraform.local/local/rhcs/0.0.1/linux_amd64
2. You now need to update your `main.tf` to the location of the local provider  by pointing the required_providers rhcs to the local terraform directory.

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

### Testing binary
If you want to locally test the provider binary without building from sources, you can pull the `latest` container image and copy the binary from the directory :
    
    <HOME>/.terraform.d/plugins/terraform.local/local/rhcs/<VERSION>/<TARGET_ARCH>
    
to your local using the following example: 

    podman run --pull=always --rm registry.ci.openshift.org/ci/rhcs-tf-e2e:latest cat /root/.terraform.d/plugins/terraform.local/local/rhcs/1.0.1/linux_amd64/terraform-provider-rhcs > ~/terraform-provider-rhcs && chmod +x ~/terraform-provider-rhcs
    can also use specific commit images by substituting `latest` for the desired commit SHA.
