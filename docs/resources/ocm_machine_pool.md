---
page_title: "ocm_machine_pool Resource"
subcategory: ""
description: |-
  Machine pool.
---

# ocm_machine_pool (Resource)

Machine pool.

## Schema

### Required

- **cluster_id** (String) Identifier of the cluster.

- **machine_type** (String) Identifier of the machine type used by the nodes,
  for example `r5.xlarge`. Use the `ocm_machine_types` data source to find the
  possible values.

- **name** (String) Name of the machine pool.

- **replicas** (Number) The number of machines of the pool

### Read-Only

- **id** (String) Unique identifier of the machine pool.
