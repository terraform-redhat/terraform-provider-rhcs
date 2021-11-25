---
page_title: "ocm_cloud_providers Data Source"
subcategory: ""
description: |-
  List of cloud providers.
---

# ocm_cloud_providers (Data Source***REMOVED***

List of cloud providers.

## Schema

### Optional

- **order** (String***REMOVED*** Order criteria.

  The syntax of this parameter is similar to the syntax of the _order by_ clause
  of a SQL statement, but using the names of the attributes of the cloud
  provider instead of the names of the columns of a table. For example, in order
  to sort the clusters descending by name identifier the value should be:

  ```sql
  name desc
  ```

  If the parameter isn't provided, or if the value is empty, then the order of
  the results is undefined.

- **search** (String***REMOVED*** Search criteria.

  The syntax of this parameter is similar to the syntax of the _where_ clause
  of a SQL statement, but using the names of the attributes of the cloud
  provider instead of the names of the columns of a table. For example, in
  order to retrieve all the cloud providers with a name starting with `A` the
  value should be:

  ```sql
  name like 'A%'
  ```

  If the parameter isn't provided, or if the value is empty, then all the
  cloud providers will be returned.

### Read-Only

- **item** (Attributes***REMOVED*** Content of the list when it has exactly one item. (see
  [below for nested schema](#nestedatt--items***REMOVED******REMOVED***

  This is intended to simplify use of the results in typical use cases, like
  searching for a specific cloud provider. For example, to search for the
  `AWS` cloud provider and then use it to create a cluster:

  ```hcl
  data "ocm_cloud_providers" "aws" {
    search = "display_name = 'AWS'"
  }

  resource "ocm_cluster" "my_cluster" {
    name           = "my-cluster"
    cloud_provider = data.ocm_cloud_providers.aws.item.id
    ...
  }
  ```

  You can also use the `items` attribute, which is always populated, but it is
  more verbose because it requires converting the result with the `tolist`
  function:

  ```hcl
  data "ocm_cloud_providers" "aws" {
    search = "display_name = 'AWS'"
  }

  resource "ocm_cluster" "my_cluster" {
    name           = "my-cluster"
    cloud_provider = tolist(data.ocm_cloud_providers.aws.items***REMOVED***[0].id
    ...
  }
  ```

- **items** (Attributes List***REMOVED*** Content of the list. (see [below for nested
  schema](#nestedatt--items***REMOVED******REMOVED***

<a id="nestedatt--items"></a>
### Nested Schema for `item` and `items`

Read-Only:

- **display_name** (String***REMOVED*** Human friendly name of the cloud provider, for
  example `AWS` or `GCP`.

- **id** (String***REMOVED*** Unique identifier of the cloud provider. This is what should
  be used when referencing the cloud providers from other places, for example in
  the `cloud_provider` attribute of the cluster resource.

- **name** (String***REMOVED*** Short name of the cloud provider, for example `aws` or `gcp`.