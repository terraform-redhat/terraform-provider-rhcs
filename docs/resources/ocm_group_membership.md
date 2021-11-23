---
page_title: "ocm_group_membership Resource"
subcategory: ""
description: |-
  Manages group membership.
---

# ocm_group_membership (Resource)

This resource manages group membership. For example, to add the user `my-user`
to the `dedicated-admins` group of a cluster:

```hcl
resource "ocm_group_membership" "my_admin" {
  cluster = ocm_cluster.my_cluster.id
  group   = "dedicated-admins"
  user    = "my-user"
}
```

Note that this will only add the user to the group, it will not create the user.
To create users use the `ocm_identity_provider` resource to create an identity
provider for the cluster and pupulate that identity provider with the users you
need.

## Schema

### Required

- **cluster** (String) Identifier of the cluster.

- **group** (String) Identifier of the group, for example `dedicated-admins`.

- **user** (String) Identifier of the user.

### Read-Only

- **id** (String) Unique identifier of the group membership.
