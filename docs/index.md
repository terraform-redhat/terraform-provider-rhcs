---
page_title: "OCM Provider"
subcategory: "Cloud Automation"
description: |-
  Experimental provider for creating and managing OpenShift managed clusters
  using the OpenShift Cluster Manager application programming interface.
---

# OCM Provider

> **IMPORTANT**: The version of the provider is currently 0.1 to indicate that
> it is at very early stage of development. The functionality isn't complete
> and there is no backwards compatibility guarantee.
>
> When it is ready for production the version will be updated to 1.0.

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
      source  = "rh-mobb/ocm"
    }
  }
}

provider "ocm" {
  token = "..."
}

resource "ocm_cluster" "my_cluster" {
  name           = "my-cluster"
  cloud_provider = "aws"
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
If this attribute isn't used then the provider will try to get the token from
the `OCM_TOKEN` environment variable.

## Schema

### Optional

- **client_id** (String) OpenID client identifier.

- **client_secret** (String, Sensitive) OpenID client secret.

- **insecure** (Boolean) When set to `true` enables insecure communication
  with the server. This disables verification of TLS certificates and host names
  and it isn't recommended for production environments. The default value is
  `false`.

- **token** (String, Sensitive) Access or refresh token. If this isn't
  explicitly provided and o other mechanism to obtain credentials is used
  (client identifier and secret) then the value will be take from the
  `OCM_TOKEN` environment variable, if that exists.

- **token_url** (String) OpenID token URL. The default is to use the _Red Hat_
  single sing on service, and there is usually no need to change it.

- **trusted_cas** (String) PEM encoded certificates of authorities that will
  be trusted. If this isn't explicitly specified then the provider will trust
  the certificate authorities trusted by default by the system.

- **url** (String) URL of the API server. If this ins't explicitly provided
  then the value will be taken from the `OCM_URL` environment variable. The
  default value is `https://api.openshift.com`.