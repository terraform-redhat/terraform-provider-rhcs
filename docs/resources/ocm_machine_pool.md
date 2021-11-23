---
page_title: "ocm_machine_pool Resource"
subcategory: ""
description: |-
  Machine pool.
---

# ocm_machine_pool (Resource***REMOVED***

Machine pool.

## Schema

### Required

- **cluster** (String***REMOVED*** Identifier of the cluster.

- **machine_type** (String***REMOVED*** Identifier of the machine type used by the nodes,
  for example `r5.xlarge`. Use the `ocm_machine_types` data source to find the
  possible values.

- **name** (String***REMOVED*** Name of the machine pool.

- **replicas** (Number***REMOVED*** The number of machines of the pool

### Read-Only

- **id** (String***REMOVED*** Unique identifier of the machine pool.
