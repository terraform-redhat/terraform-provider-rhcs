<a href="https://redhat.com">
    <img src=".github/Logo_Red_Hat.png" alt="RedHat logo" title="RedHat" align="right" height="50" />
</a>

# Red Hat OCM Provider

> Please note that this Terraform provider and its modules are open source and will continue to iterate features, gradually maturing this code.
> If you encounter any issues, please report them in this repo.

## Introduction

The OCM provider allows Terraform to manage Red Hat OpenShift Service on AWS (ROSA***REMOVED*** clusters, machine pools and identity provider.

## OCM

Red Hat OpenShift Cluster Manager is a managed service where you can install, modify, operate, and upgrade your Red Hat OpenShift clusters. This service allows you to work with all of your organization’s clusters from a single dashboard. More information can be found [here](https://docs.openshift.com/rosa/ocm/ocm-overview.html***REMOVED***.

## ROSA
Red Hat OpenShift Service on AWS (ROSA***REMOVED*** is a fully-managed, turnkey application platform that allows you to focus on delivering value to your customers by building and deploying applications. 
More information can be found [here](https://docs.openshift.com/rosa/welcome/index.html***REMOVED***.

## AWS STS
A Secure Token Service (STS***REMOVED*** is a component that issues, validates, renews, and cancels security tokens for trusted systems, users, and resources requesting access within a federation.
AWS provides AWS STS as a web service that enables you to request temporary, limited-privilege credentials for AWS Identity and Access Management (IAM***REMOVED*** users or for users you authenticate (federated users***REMOVED***.

## ROSA STS mode

To deploy a Red Hat OpenShift Service on AWS (ROSA***REMOVED*** cluster that uses the AWS Security Token Service (STS***REMOVED***, you must create the following AWS Identity Access Management (IAM***REMOVED*** resources:

Specific account-wide IAM roles and policies that provide the STS permissions required for ROSA support, installation, control plane, and compute functionality. This includes account-wide Operator policies.
Cluster-specific Operator IAM roles that permit the ROSA cluster Operators to carry out core OpenShift functionality.
An OpenID Connect (OIDC***REMOVED*** provider that the cluster Operators use to authenticate.

## Terraform

Terraform is an infrastructure as a code tool, used primarily by DevOps teams.
It lets you define resources in human-readable configuration files that you can version, reuse, and share.

Terraform creates and manages resources through their application programming interfaces (APIs***REMOVED*** by using "Providers".

## Prerequisites

In order to use the provider inside your terraform configuration you need to import it using:

* Offline token (OCM***REMOVED***:

Get an offline token from [https://console.redhat.com/openshift/token/rosa](https://console.redhat.com/openshift/token/rosa***REMOVED***

* Create ROSA account IAM roles: 

Detailed ROSA Account Roles and Policies can be found [here](https://docs.openshift.com/rosa/rosa_architecture/rosa-sts-about-iam-resources.html***REMOVED***

```
rosa create account-roles
```

* Least AWS Permissions required to run the terraform

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "iam:GetRole",
                "iam:UpdateOpenIDConnectProviderThumbprint",
                "iam:CreateRole",
                "iam:DeleteRole",
                "iam:UpdateRole",
                "iam:DeleteOpenIDConnectProvider",
                "iam:GetOpenIDConnectProvider",
                "iam:CreateOpenIDConnectProvider",
                "iam:TagOpenIDConnectProvider",
                "iam:TagRole",
                "iam:ListRolePolicies",
                "iam:ListAttachedRolePolicies",
                "iam:ListInstanceProfilesForRole",
                "iam:AttachRolePolicy",
                "iam:DetachRolePolicy"
            ],
            "Resource": [
                "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/*",
                "arn:aws:iam::<ACCOUNT_ID>:role/*"
            ]
        }
    ]
}
```

* Choose operator IAM roles prefix name

The operator IAM roles will be created per cluster by [module](https://registry.terraform.io/modules/terraform-redhat/rosa-sts***REMOVED***.


## Sample Terraform Manifest File

```
variable token {
    type = string
}

variable operator_role_prefix {
    type = string
}

variable cluster_name {
    type = string
}

variable region {
    type = string
}

variable zone {
    type = string
}

provider "ocm" {
  token = var.token
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.20.0"
    }
    ocm = {
      version = ">= 1.0"
      source  = "terraform-redhat/ocm"
    }
  }
}

locals {
  sts_roles = {
      role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Installer-Role",
      support_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Support-Role",
      instance_iam_roles = {
        master_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-ControlPlane-Role",
        worker_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ManagedOpenShift-Worker-Role"
      },
      operator_role_prefix = var.operator_role_prefix,
  }
}

data "aws_caller_identity" "current" {}

resource "ocm_cluster_rosa_classic" "rosa_sts_cluster" {
  name           = var.cluster_name
  cloud_region   = var.region
  aws_account_id     = data.aws_caller_identity.current.account_id
  availability_zones = [var.zone]
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  sts = local.sts_roles
}

data "ocm_rosa_operator_roles" "operator_roles" {
  operator_role_prefix = var.operator_role_prefix
}

module operator_roles {
    source = "terraform-redhat/rosa-sts/aws"
    version = "0.0.1"
    cluster_id = ocm_cluster_rosa_classic.rosa_sts_cluster.id
    rh_oidc_provider_thumbprint = ocm_cluster_rosa_classic.rosa_sts_cluster.sts.thumbprint
    rh_oidc_provider_url = ocm_cluster_rosa_classic.rosa_sts_cluster.sts.oidc_endpoint_url
    operator_roles_properties = data.ocm_rosa_operator_roles.operator_roles.operator_iam_roles
}
```

## Development Introduction
Running `terraform plan` against a local build of OCM provider can be done by those steps:
1. Run  ```make install ```. After running ```make install``` you will find the ocm provider binary file in the directory: 
```
<HOME>/.terraform.d/plugins/terraform.local/local/ocm/<VERSION>/<TARGET_ARCH>
```

for example 
```
~/.terraform.d/plugins/terraform.local/local/ocm/0.0.1/linux_amd64
```

2. Point to the local provider by pointing the required_providers ocm to the local terraform directory

```
terraform {
  required_providers {
    ocm = {
      source  = "terraform.local/local/ocm"
      version = "0.0.1"
    }
  }
}


provider "ocm" {
  token = var.token
}
```
