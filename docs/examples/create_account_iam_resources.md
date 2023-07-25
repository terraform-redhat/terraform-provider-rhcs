# Account-wide IAM roles

Prior to creating a ROSA STS cluster, you must create the required account-wide roles and policies. For more information, see [Account-wide IAM role and policy reference](https://access.redhat.com/documentation/en-us/red_hat_openshift_service_on_aws/4/html/introduction_to_rosa/rosa-sts-about-iam-resources#rosa-sts-account-wide-roles-and-policies_rosa-sts-about-iam-resources) in the Red Hat Customer Portal.
Those account-wide AWS resources can be created only ones per account and used for all Rosa clusters creations in this account.

## Prerequisites

To follow this tutorial you will need:

* The [Terraform CLI](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli) (1.2.0+) installed.
* The [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) installed.
* [AWS account](https://aws.amazon.com/free/?all-free-tier) and [associated credentials](https://docs.aws.amazon.com/IAM/latest/UserGuide/security-creds.html) that allow you to create resources. The credentials configured for the AWS provider (see [Authentication and Configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration) section in AWS terraform provider documentations)
* The [ROSA CLI](https://docs.openshift.com/rosa/rosa_cli/rosa-get-started-cli.html) installed
* [OpenShift Cluster Manager API Token](https://console.redhat.com/openshift/token)

~> **NOTE:** In this example, the terraform module [rosa-sts](https://registry.terraform.io/modules/terraform-redhat/rosa-sts/aws/latest) is utilized, encompassing various AWS resources essential for creating and managing AWS IAM resources that are employed for Rosa cluster actions.

~> **NOTE:** In this example, beside the IAM roles and policies which used for the ROSA resources creation, additional policies are created to be used by the IAM roles would be created for the ROSA cluster and would be used by the ROSA cluster operators 

~> **NOTE:** This example uses "Red Hat Cloud Services Provider" (this provider) to get a list of all OCP supported versions, this list need to be added to the "rosa-sts" module apply, in order to validate the requested version. Also it used to get the list of all permissions required for policy creation.

## Input Variables (variables.tf)

This example include some variables which need to be added before running `terraform apply`
Terraform requires a value for every variable. There are several ways to assign variable values. see [Assigning Values to Root Module Variables](https://developer.hashicorp.com/terraform/language/values/variables#assigning-values-to-root-module-variables) for more info.

```terraform
variable "ocm_environment" {
  type    = string
  default = "production"
}

variable "openshift_version" {
  type = string
  default = ""
}

variable "account_role_prefix" {
  type    = string
  default = ""
}

variable "token" {
  type = string
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}

variable "path" {
  description = "(Optional) The arn path for the account/operator roles as well as their policies."
  type        = string
  default     = null
}

variable "tags" {
  description = "List of AWS resource tags to apply."
  type        = map(string)
  default     = null
}
```

## Terraform output (output.tf)

```terraform
output "account_role_prefix" {
  value = module.create_account_roles.account_role_prefix
}
```

## Provider decleration (main.tf)

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    rhcs = {
      source  = "terraform-redhat/rhcs"
    }
  }
}

provider "rhcs" {
  token = var.token
  url   = var.url
}
```

## Usage example (iam_resources.tf)

```terraform
data "rhcs_policies" "all_policies" {}

data "rhcs_versions" "all" {}

module "create_account_roles" {
  source  = "terraform-redhat/rosa-sts/aws"
  version = "0.0.12"

  create_operator_roles = false
  create_oidc_provider  = false
  create_account_roles  = true

  account_role_prefix    = var.account_role_prefix
  ocm_environment        = var.ocm_environment
  rosa_openshift_version = var.openshift_version
  account_role_policies  = data.rhcs_policies.all_policies.account_role_policies
  operator_role_policies = data.rhcs_policies.all_policies.operator_role_policies
  all_versions           = data.rhcs_versions.all
  path                   = var.path
  tags                   = var.tags
}
```

## Verification

1. In the `rosa` CLI, run the following command to verify that the account roles are created:
    ````
    rosa list account-roles
    ````
1. You see your roles when the command finishes. 
    ````
    I: Fetching account roles
    ROLE NAME                           ROLE TYPE      ROLE ARN                                                    OPENSHIFT VERSION  AWS Managed
    ManagedOpenShift-ControlPlane-Role  Control plane  arn:aws:iam::XXXXX:role/ManagedOpenShift-ControlPlane-Role  4.13               No
    ManagedOpenShift-Installer-Role     Installer      arn:aws:iam::XXXXX:role/ManagedOpenShift-Installer-Role     4.13               No
    ManagedOpenShift-Support-Role       Support        arn:aws:iam::XXXXX:role/ManagedOpenShift-Support-Role       4.13               No
    ManagedOpenShift-Worker-Role        Worker         arn:aws:iam::XXXXX:role/ManagedOpenShift-Worker-Role        4.13               No

## Resource clean up

After you are done with the resources you created, you should not delete them manually, but instead, use the `destroy` command. Run the following to delete all of your created resources:
  
    terraform destroy

After the command is complete, your resources are deleted.
