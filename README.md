# Terraform provider for OCM

> **IMPORTANT**: The version of the provider is currently 0.1 to indicate that
> it is at very early stage of development. The functionality isn't complete
> and there is no backwards compatibility guarantee.
>
> When it is ready for production the version will be updated to 1.0.

## Introduction

### Create OSD AWS Cluster

The OCM provider simplifies the provisioning of _OpenShift_ managed clusters
using the [OpenShift Cluster Manager](https://console.redhat.com/openshift)
application programming interface.

For example, to create a simple cluster with an identity provider that allows
login with a simple user name and password create a `main.tf` file similar this
and then run `terraform apply`:

```hcl
terraform {
  required_providers {
    ocm = {
      version = ">= 0.1"
      source  = "openshift-online/ocm"
    }
  }
}

provider "ocm" {
  token = "..."
}

resource "ocm_cluster" "my_cluster" {
  name           = "my-cluster"
  cloud_provider = "aws"
  product        = "osd"
  cloud_region   = "us-east-1"
}

resource "ocm_identity_provider" "my_idp" {
  cluster = ocm_cluster.my_cluster.id
  name    = "my-idp"
  htpasswd = {
    username = "admin"
    password = "redhat123"
  }
}

resource "ocm_group_membership" "my_admin" {
  cluster = ocm_cluster.my_cluster.id
  group   = "dedicated-admins"
  user    = "admin"
}
```

The value of the `token` attribute of the provider should be the OCM
authentication token that you can get [here](https://console.redhat.com/openshift/token).
If this attribute isn't used then the provider will try to get the token it from
the `OCM_TOKEN` environment variable.

### Create AWS Rosa STS Cluster

The following example shows a production grade rosa cluster with:

* Existing VPC & Subnets
* Multi AZ
* Proxy
* STS

```
locals {
  sts_roles = {
      role_arn = "arn:aws:iam::${var.account_id}:role/ManagedOpenShift-Installer-Role",
      support_role_arn = "arn:aws:iam::${var.account_id}:role/ManagedOpenShift-Support-Role",
      operator_iam_roles = [
        {
          name =  "cloud-credential-operator-iam-ro-creds",
          namespace = "openshift-cloud-credential-operator",
          role_arn = "arn:aws:iam::${var.account_id}:role/${var.operator_role_prefix}-openshift-cloud-credential-operator-cloud-c",
        },
        {
          name =  "installer-cloud-credentials",
          namespace = "openshift-image-registry",
          role_arn = "arn:aws:iam::${var.account_id}:role/${var.operator_role_prefix}-openshift-image-registry-installer-cloud-cr",
        },
        {
          name =  "cloud-credentials",
          namespace = "openshift-ingress-operator",
          role_arn = "arn:aws:iam::${var.account_id}:role/${var.operator_role_prefix}-openshift-ingress-operator-cloud-credential",
        },
        {
          name =  "ebs-cloud-credentials",
          namespace = "openshift-cluster-csi-drivers",
          role_arn = "arn:aws:iam::${var.account_id}:role/${var.operator_role_prefix}-openshift-cluster-csi-drivers-ebs-cloud-cre",
        },
        {
          name =  "cloud-credentials",
          namespace = "openshift-cloud-network-config-controller",
          role_arn = "arn:aws:iam::${var.account_id}:role/${var.operator_role_prefix}-openshift-cloud-network-config-controller-c",
        },
        {
          name =  "aws-cloud-credentials",
          namespace = "openshift-machine-api",
          role_arn = "arn:aws:iam::${var.account_id}:role/${var.operator_role_prefix}-openshift-machine-api-aws-cloud-credentials",
        },
      ]
      instance_iam_roles = {
        master_role_arn = "arn:aws:iam::${var.account_id}:role/ManagedOpenShift-ControlPlane-Role",
        worker_role_arn = "arn:aws:iam::${var.account_id}:role/ManagedOpenShift-Worker-Role"
      },    
  }
}
resource "ocm_cluster" "rosa_cluster" {
  name           = var.cluster_name
  cloud_provider = "aws"
  cloud_region   = "us-east-2"
  product        = "rosa"
  aws_account_id     = "var.account_id"
  aws_subnet_ids = module.openshift_vpc.rosa_subnet_ids
  machine_cidr = module.openshift_vpc.rosa_vpc_cidr
  multi_az = true
  aws_private_link = true
  availability_zones = ["us-east-2a", "us-east-2b", "us-east-2c"]
  proxy = {
    http_proxy = var.proxy
    https_proxy = var.proxy
  }
  properties = {
    rosa_creator_arn = data.aws_caller_identity.current.arn
  }
  wait = false
  sts = local.sts_roles
}
```


## Documentation

The reference documentation of the provider is available in the Terraform
[registry](https://registry.terraform.io/providers/openshift-online/ocm/latest/docs).

## Examples

Check the [examples](examples) directory for complete examples.

## Development

To build the provider run the `make` command:

```shell
$ make
```

This will create a local Terraform plugin registry in the directory
`.terraform.d/plugins` of the project. Assuming that you have the project
checked out in `/files/projects/terraform-provider-ocm/repository` you will need
to add something like this to your Terraform CLI configuration file:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/files/projects/terraform-provider-ocm/repository/.terraform.d/plugins"
    include = ["localhost/*/*"]
  }
}
```

If you don't want to change your global CLI configuation file you can put this
in any file you like and then use the `TF_CLI_CONFIG_FILE` environment variable
to point to it. For example, put the configuration in
`/files/projects/terraform-provider-ocm/terraform.rc` and then set the
environment variable pointing to it:

```shell
$ cat >/files/projects/terraform-provider-ocm/terraform.rc <<.
provider_installation {
  filesystem_mirror {
    path    = "/files/projects/terraform-provider-ocm/repository/.terraform.d/plugins"
    include = ["localhost/*/*"]
  }
}
.
$ export TF_CLI_CONFIG_FILE=/files/projects/terraform-provider-ocm/terraform.rc
```

Once your configuration is ready you can go to the directory containing the
Terraform `.tf` files and run the `terraform init` and `terraform apply`
commands:

```shell
$ terraform init
$ terraform apply
```

To see the debug log of the provider set the `TF_LOG` environment variable to
`DEBUG` before running the `terraform apply` command:

```shell
$ export TF_LOG=DEBUG
$ terraform apply
```
