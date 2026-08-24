---
page_title: "Default Machine Pool in ROSA Cluster"
subcategory: ""
description: |-
  Guide explaining the default machine pool generated as a component of the ROSA cluster.
---

# Default Machine Pool

## Introduction

Upon the creation of a new ROSA cluster, default Machine Pool(s) are automatically generated. This step is essential as the cluster cannot attain a READY state until its worker nodes are operational.

### Classic clusters

Classic ROSA clusters create a single default Machine Pool named `worker`. Users can configure certain properties of the default Machine Pool by adjusting the corresponding attributes within the `rhcs_cluster_rosa_classic` resource (for more info about the attributes see [ROSA Cluster attributes list](../resources/cluster_rosa_classic.md)).

### HCP clusters

ROSA HCP clusters create default Machine Pool(s) automatically when the cluster is created:

- **Single-AZ clusters:** one pool named `workers`
- **Multi-AZ clusters:** one pool per availability zone, named `workers-0`, `workers-1`, `workers-2`, and so on

For HCP clusters, each default Machine Pool requires a `subnet_id` in the `rhcs_hcp_machine_pool` resource. The subnet maps to the availability zone at the same index in the cluster's `aws_subnet_ids` list (for example, `workers-0` uses the first subnet, `workers-1` the second, and so on).

Following the creation of the cluster, the default Machine Pool attributes used during cluster creation (such as `desired_capacity`, `instance_type`, etc.) are retained in the cluster resource state and become unchangeable. To make any changes to the default Machine Pool or to delete it, you must first import a Machine Pool resource pointing to this default pool, after which changes can be made through the imported resource.

## Import the default Machine Pool resource

Users can choose from two methods to import the default Machine Pool:
### Option 1: terraform import command
After creating the cluster, users can incorporate the relevant resource by utilizing the terraform import command.

### Option 2: "Magic import"
The resource can be included in the manifest at any stage (including the same manifest where the ROSA cluster is declared, before applying). Subsequently, executing terraform apply will trigger a unique behavior specifically designed for importing the Default Machine Pool. Magic import applies to all default Machine Pool names: `worker` for Classic, `workers` for single-AZ HCP clusters, and `workers-0`, `workers-1`, `workers-2`, and so on for multi-AZ HCP clusters.
> Note: Using the magic import could result in optional attributes being overwritten (i.e. labels, taints, replicas, max_replicas, min_replicas... etc)

## Limitations

* Default Machine Pool names are reserved and cannot be used for other machine pools:
  * Classic: `worker`
  * HCP single-AZ: `workers`
  * HCP multi-AZ: `workers-0`, `workers-1`, `workers-2`, and so on
* The special import flow during the apply process is only applicable to these default Machine Pool names.
* Every ROSA Cluster must include at least one Machine Pool to meet the cluster's minimal node requirement. Consequently, deleting the last Machine Pool will only involve removing the Terraform resource and not deleting the actual resource in the backend.
