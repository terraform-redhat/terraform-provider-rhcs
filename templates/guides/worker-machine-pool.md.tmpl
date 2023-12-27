---
page_title: "Default Machine Pool in Rosa Cluster"
subcategory: "Description on default machine pool, created as part of rosa cluster"
description: |-
  .
---

# Default Machine Pool

## Introduction

Upon the creation of a new Rosa cluster, a default Machine Pool named "worker" is automatically generated. This step is essential as the cluster cannot attain a READY state until its worker nodes are operational. Users have the flexibility to configure certain properties of the default Machine Pool by adjusting the corresponding attributes within the Rosa cluster resource (For more info about the attributes see [Rosa Cluster attributes list](../resources/machine_pool.md).
Following the creation of the cluster, the attributes in the Rosa cluster resource lose their relevance. The values utilized during cluster creation are retained in the state, becoming unchangeable. These values may not accurately reflect the current settings in the backend resource.
In order to make any change in the default Machine Pool or to delete it, user must import a Machine Pool resource pointing to this default resource first, than any change can be done on this resource.

## Import the default Machine Pool resource


Users can choose from two methods to import the default Machine Pool:
* After creating the cluster, users can incorporate the relevant resource by utilizing the terraform import command.
* The resource can be included in the manifest at any stage (including the same manifest where the Rosa cluster is declared, before applying). Subsequently, executing terraform apply will trigger a unique behavior specifically designed for importing the Default Machine Pool, with a focus on the resource named "worker."

## Limitations

* The name "worker" for the Machine Pool is exclusively reserved for the default Machine Pool, and no other Machine Pool can be created with this particular name.
* The special import flow during the apply process is only applicable to the Default Machine Pool named "worker."
* Every Rosa Cluster must include at least one Machine Pool to meet the cluster's minimal node requirement. Consequently, deleting the last Machine Pool will only involve removing the Terraform resource and not deleting the actual resource in the backend.
