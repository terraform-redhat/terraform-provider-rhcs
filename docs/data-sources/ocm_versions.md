---
page_title: "ocm_versions Data Source"
subcategory: ""
description: |-
  List of OpenShift versions.
---

# ocm_versions (Data Source***REMOVED***

This data source lists the _OpenShift_ verions that can be used to create
clusters.

## Schema

### Read-Only

- **items** (Attributes List***REMOVED*** Items of the list. (see [below for nested
  schema](#nestedatt--items***REMOVED******REMOVED***

<a id="nestedatt--items"></a>
### Nested Schema for `items`

Read-Only:

- **id** (String***REMOVED*** Identifier of the version. This is what should be used when
  referencing the version from other places, for example in the `version`
  attribute of the cluster resource.

- **name** (String***REMOVED*** Short name of the the version, for example `4.1.0`.