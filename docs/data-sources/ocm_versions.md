---
page_title: "ocm_versions Data Source"
subcategory: ""
description: |-
  List of OpenShift versions.
---

# ocm_versions (Data Source)

This data source lists the _OpenShift_ verions that can be used to create
clusters.

## Schema

- **order** (String) Order criteria.

  The syntax of this parameter is similar to the syntax of the _order by_ clause
  of a SQL statement, but using the names of the attributes of the version
  instead of the names of the columns of a table. For example, in order to sort
  the versions descending by identifier the value should be:

  ```sql
  id desc
  ```

  If the parameter isn't provided, or if the value is empty, then the order of
  the results is undefined.

- **search** (String) Search criteria.

  The syntax of this parameter is similar to the syntax of the _where_ clause of
  a SQL statement, but using the names of the attributes of the version instead
  of the names of the columns of a table. For example, in order to retrieve all
  the versions from the _fast_ channel group:

  ```sql
  enabled = 't' and channel_group = 'fast'
  ```

  Note that the default is to search for enabled versions, equivalent to
  `enabled = 't'`. If you set this attribute that default value will be
  overriden. Make sure to add it to your search criteria unless you also want
  the versions that are disabled.

### Read-Only

- **item** (Attributes) Content of the list when it has exactly one item. (see
  [below for nested schema](#nestedatt--items))

  This is intended to simplify use of the results in typical use cases, like
  searching for a specific version. For example, to search for the versions in
  the `fast` channel group and use the first one:

  ```hcl
  data "ocm_versions" "fast" {
    search = "enabled = 't' and channel_group = 'fast'"
    order  = "id asc"
  }

  resource "ocm_cluster" "my_cluster" {
    name    = "my-cluster"
    version = data.ocm_versions.fast.item.id
    ...
  }
  ```

  You can also use the `items` attribute, which is always populated, but it is
  more verbose because it requires converting the result with the `tolist`
  function:

  ```hcl
  data "ocm_versions" "fast" {
    search = "enabled = 't' and channel_group = 'fast'"
    order  = "id asc"
  }

  resource "ocm_cluster" "my_cluster" {
    name    = "my-cluster"
    version = tolist(data.ocm_versions.fast.items)[0].id
    ...
  }
  ```
- **items** (Attributes List) Items of the list. (see [below for nested
  schema](#nestedatt--items))

<a id="nestedatt--items"></a>
### Nested Schema for `items`

Read-Only:

- **id** (String) Identifier of the version. This is what should be used when
  referencing the version from other places, for example in the `version`
  attribute of the cluster resource.

- **name** (String) Short name of the the version, for example `4.1.0`.