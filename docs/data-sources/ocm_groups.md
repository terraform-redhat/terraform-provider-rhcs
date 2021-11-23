---
page_title: "ocm_groups Data Source"
subcategory: ""
description: |-
  List of groups.
---

# ocm_groups (Data Source)

Lists the groups of users of a cluster.

Currently there is only one group named `dedicated-admins` is supported, so
this data source will always return exactly one item.

## Schema

### Required

- **cluster** (String) Identifier of the cluster.

### Read-Only

- **items** (Attributes List) Items of the list. (see [below for nested schema](#nestedatt--items))

<a id="nestedatt--items"></a>
### Nested Schema for `items`

Read-Only:

- **id** (String) Unique identifier of the group. This is what should be used
  when referencing the group from other places, for example in the `group`
  attribute of the user resource.

- **name** (String) Short name of the group for example `dedicated-admins`.


