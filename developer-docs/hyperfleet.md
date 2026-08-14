# Hyperfleet resources

This document covers the architecture, file layout, schema design, and test strategy for the two Hyperfleet resources (`rhcs_cluster_hyperfleet` and `rhcs_nodepool_hyperfleet`). It also records the deliberate gaps relative to their HCP equivalents and the reasons behind them.

## What is Hyperfleet

Hyperfleet exposes a Kubernetes-native Platform API (v2) for managing ROSA HCP clusters. It is distinct from OCM in several important ways:

| Concern | OCM (HCP) | Hyperfleet |
|---|---|---|
| Authentication | OCM offline token | Ambient AWS credentials (env vars, shared config, instance profile) |
| API style | OCM REST SDK | Kubernetes-style clientset (Get/List/Create/Update/Delete/Watch) |
| Resource model | OCM resources (clusters, machine pools) | Kubernetes custom resources (`Cluster`, `NodePool` in `hyperfleet.io/v1alpha1`) |
| HyperShift fields | Mediated via OCM API model | Passed through to HyperShift types directly (`HostedClusterSpecPassthrough`, `NodePoolSpecPassthrough`) |
| Operator roles | Managed by OCM IAM module | User-managed; constructed from `operator_roles_prefix` and account ID |
| OIDC provider | Created by OCM | Created by the user after cluster creation (IAM tf-manifest) |

Because Hyperfleet resources share no auth or API surface with OCM, they cannot use the OCM SDK client (`cmv1`) and cannot participate in OCM-level features (channels, machine pool autoscaling, tuning configs, kubelet configs, etc.).

## Architecture decisions

### No OCM token; provider config is different
The provider `rhcs` block for Hyperfleet resources requires `hyperfleet_url` and `aws_account_id`. `aws_account_id` is sent as the mandatory `X-Amz-Account-Id` SigV4 signed header. `aws_caller_arn` is optional: when supplied it is added as the `X-Amz-Caller-Arn` signed header (and becomes the cluster `CreatorARN`); when omitted, requests are still signed and authenticated with the ambient AWS credentials and account ID. `aws_region` is also optional — it is derived from the `hyperfleet_url` execute-api hostname when unset. No `token` or `url` (OCM gateway) is needed. The client is `hyperfleet.Interface` from `rosa-hyperfleet-api/clientset`, not `*cmv1.Client`.

### Passthrough write-mode governs what users can set
The Hyperfleet API annotates every HyperShift field with `+hyperfleet:write-mode`:
- `mutable` — users can set and update the field.
- `service-set` — the platform operator owns the field; user writes are rejected.

Fields marked `service-set` cannot be exposed as writable resource attributes. See the gap tables below for the full breakdown.

### Subsystem test layer covers provider-config diagnostics only
The standard `subsystem/hcp/` and `subsystem/classic/` suites rely on a stubbed OCM HTTP server (`TestServer`). No equivalent stub exists for the Hyperfleet Platform API, which uses a Kubernetes API server wire protocol, so the subsystem suite cannot exercise the CRUD path. The `subsystem/hyperfleet/` suite therefore covers only the provider-configuration diagnostics that fire before any Platform API call:
- `hyperfleet_url` absent from the provider block → descriptive error.
- `hyperfleet_url` set but `aws_account_id` missing → descriptive error.
- `aws_region` contradicting the region in `hyperfleet_url` → "AWS region mismatch" error.

Full field-mapping and CRUD coverage is provided by:
- Unit tests in `provider/hyperfleet/*_test.go` (state mapping, `buildNodePoolSpec`, `populateState`).
- E2E tests in `tests/e2e/hyperfleet/` against a real Hyperfleet environment.

### Worker instance profile is computed, not exposed
The Hyperfleet cluster object carries the 7 operator role ARNs (`RolesRef`) but **not** the worker IAM instance profile. If the nodepool spec omits it, CAPA defaults to a non-existent `<infra-id>-worker-profile` and the create fails with `Invalid IAM Instance Profile name`. The provider therefore computes it during `Create`: it fetches the parent cluster, recovers the operator roles prefix from `RolesRef` (via `prefixAndPartitionFromRolesRef`), and sets `spec.NodePool.Platform.AWS.InstanceProfile = <prefix>-ROSA-Worker-Role` — matching the IAM manifest naming convention (`tests/tf-manifests/aws/iam-roles/hyperfleet/main.tf`). On `Update` the value is preserved from the live object.

This value is **derived, not user-supplied**, and is intentionally *not* a schema attribute or state field — there is no per-nodepool customization today. If a future use case needs custom instance profiles, promote it to an optional `aws_node_pool.instance_profile` attribute and read it back from the API response in `populateNodePoolState` instead of computing it (see the note in `nodepool_resource.go` `Create`).

### Separate nodepool resource (not the default machine pool)
HCP clusters created via OCM have a default machine pool managed by `rhcs_cluster_rosa_hcp`. Hyperfleet clusters have no managed default pool — all node pools are explicit `rhcs_nodepool_hyperfleet` resources. The cluster resource exposes only control-plane configuration.

### Fire-and-forget delete for nodepools
The Hyperfleet API returns 200 for a nodepool delete request immediately; the operator drains and deprovisions asynchronously. The Terraform resource clears state on a 200 or 404. A PDB on a cluster may block eviction indefinitely — those nodepools are expected to stay in `Deleting` phase until the cluster itself is deleted (which cascades cleanup).

### Cluster deletion is synchronous in the e2e test
The e2e teardown uses `WaitUntil` (a synchronous ticker loop from `bridge_wrappers_generated.go`) to block until the cluster's 404 is confirmed before IAM and VPC teardown runs. This prevents race conditions where the operator is still releasing VPC resources (ENIs, security groups) when `terraform destroy` runs on the VPC.

### VPC destroy uses a retry loop
After the cluster object is deleted (404), the cluster operator continues releasing AWS resources asynchronously. The e2e teardown wraps `vpcSvc.Destroy()` in an `Eventually` with a 30-minute timeout and 2-minute polling interval to tolerate this delay. Classic load balancers and non-default security groups are pruned first using `helper.DeleteClassicLoadBalancers` / `helper.DeleteNonDefaultSecurityGroups` before the retry loop.

### VPC teardown purges the private hosted zone
The cluster operator writes CNAME records (`api.<name>.hypershift.local`, `*.apps.<name>.hypershift.local`) into the private hosted zone created by the VPC manifest (`aws_route53_zone.hyperfleet`). Route53 refuses to delete a non-empty zone, so `terraform destroy` of the VPC fails with `HostedZoneNotEmpty` while those records remain. The e2e teardown calls `helper.PurgeHostedZoneRecords(region, vpcOut.HostedZoneID)` before `vpcSvc.Destroy()` to delete every record except the NS/SOA pair (which Route53 manages with the zone lifecycle itself). The hosted zone ID is surfaced by the VPC manifest as the `hosted_zone_id` output.

### Cluster name length limit
The Hyperfleet operator creates a namespace named `cluster-<36-char-uuid>-<cluster-name>`. Kubernetes namespaces are capped at 63 characters. After the fixed prefix and UUID, at most 18 characters remain for the cluster name (`63 - len("cluster-") - 36 - len("-") = 18`). The e2e default name `hf-e2e-<unix-ts>` is within this limit; names beyond 18 characters cause a 422 from the API.

### Import state
Both resources implement `ResourceWithImportState`. The cluster resource imports by cluster UUID. The nodepool resource imports by `<cluster_uuid>/<nodepool_name>`.

## File layout

```
provider/hyperfleet/
├── resource.go               # rhcs_cluster_hyperfleet — Schema, CRUD, ImportState
├── state.go                  # ClusterHyperfleetState struct with tfsdk tags
├── resource_test.go          # Unit tests for populateState, computeRolesRef
├── nodepool_resource.go      # rhcs_nodepool_hyperfleet — Schema, CRUD, ImportState
├── nodepool_state.go         # NodePoolHyperfleetState, NPAWSNodePool structs
└── nodepool_resource_test.go # Unit tests for buildNodePoolSpec, populateNodePoolState

tests/e2e/hyperfleet/
├── suite_test.go             # Ginkgo suite registration (TestHyperfleetSanity)
└── sanity_test.go            # End-to-end: VPC → cluster → IAM → nodepools → scale → teardown

tests/tf-manifests/rhcs/
├── clusters/hyperfleet/      # Terraform root for rhcs_cluster_hyperfleet
│   ├── main.tf               # Provider config + resource block
│   ├── variables.tf          # Input variables
│   └── output.tf             # cluster_id, cluster_name, cluster_phase, cluster_api_url, oidc_issuer
├── node-pools/hyperfleet/    # Terraform root for rhcs_nodepool_hyperfleet
│   ├── main.tf
│   ├── variables.tf
│   └── output.tf             # nodepool_id, nodepool_name, nodepool_phase, nodepool_replicas
└── (vpc/, iam/ are under a separate hyperfleet path — see exec helpers)

tests/utils/exec/
├── hyperfleet_cluster.go     # HyperfleetClusterService + HyperfleetClusterArgs/Output
├── hyperfleet_nodepool.go    # HyperfleetNodePoolService + HyperfleetNodePoolArgs/Output
├── hyperfleet_iam.go         # HyperfleetIAMService + HyperfleetIAMArgs/Output
└── hyperfleet_vpc.go         # HyperfleetVPCService + HyperfleetVPCArgs/Output
```

The `manifests.GetHyperfleet*ManifestsDir()` helpers in `tests/utils/exec/manifests/manifests.go` map each service to its tf-manifest directory.

## Schema reference

### `rhcs_cluster_hyperfleet`

| Attribute | Type | R/O/C | Immutable | Notes |
|---|---|---|---|---|
| `id` | string | C | — | Cluster UUID (from Platform API `metadata.uid`) |
| `name` | string | R | yes | Human-readable name; max 18 chars (namespace limit) |
| `operator_roles_prefix` | string | R | yes | Prefix for the 7 operator IAM roles |
| `aws_subnet_ids` | list(string) | R | yes | Worker subnet IDs; first entry used as cluster subnet |
| `vpc_id` | string | R | yes | VPC containing the worker subnet |
| `availability_zones` | list(string) | R | yes | Worker AZs; first entry used |
| `expiration_timestamp` | string | O | no | RFC3339; platform auto-deletes cluster at this time |
| `aws_partition` | string | O/C | yes | Default `aws`; set to `aws-us-gov` for GovCloud |
| `cloud_region` | string | O/C | yes | Derived from `availability_zones` if unset |
| `creator_arn` | string | C | — | IAM ARN from provider `aws_caller_arn` |
| `phase` | string | C | — | WaitingForPlacement / Provisioning / Ready / Deleting |
| `api_url` | string | C | — | Control-plane API endpoint; empty until Ready |
| `current_version` | string | C | — | Running OpenShift version |
| `management_cluster` | string | C | — | Management cluster ID assigned by Platform API |
| `oidc_issuer` | string | C | — | OIDC issuer URL; use to create OIDC provider + trust policies |

R=Required, O=Optional, C=Computed.

### `rhcs_nodepool_hyperfleet`

| Attribute | Type | R/O/C | Immutable | Notes |
|---|---|---|---|---|
| `id` | string | C | — | NodePool UID |
| `name` | string | R | yes | |
| `cluster` | string | R | yes | Parent cluster UUID (`id` output of `rhcs_cluster_hyperfleet`) |
| `subnet_id` | string | R | yes | AWS subnet for node instances |
| `replicas` | int64 | O | no | Fixed replica count |
| `labels` | map(string) | O | no | Node labels |
| `auto_repair` | bool | R | no | |
| `aws_node_pool.instance_type` | string | R | yes | EC2 instance type |
| `aws_node_pool.disk_size` | int64 | O/C | yes | Root volume GiB |
| `aws_node_pool.tags` | map(string) | O | no | Additional AWS resource tags |
| `phase` | string | C | — | WaitingForCluster / Provisioning / Ready / Deleting |
| `ignore_deletion_error` | bool | O/C | no | Suppress delete errors; default false |

Import format: `<cluster_uuid>/<nodepool_name>`

## HCP parity gap analysis

### `rhcs_cluster_hyperfleet` vs `rhcs_cluster_rosa_hcp`

#### Fields that CAN be added (mutable in passthrough)

| HCP attribute | Passthrough location | Notes |
|---|---|---|
| `tags` | `ClusterSpec.Tags` | Feature-gated on `HyperFleetAutoScaling`; max 100 entries |
| `properties` | `ClusterSpec.Properties` | Arbitrary customer metadata; max 100 entries |
| `delete_protection` | `ClusterSpec.DeleteProtection` | |
| `pod_cidr` | `HostedCluster.Networking.PodCIDR` | Immutable after creation |
| `service_cidr` | `HostedCluster.Networking.ServiceCIDR` | Immutable after creation |
| `machine_cidr` | `HostedCluster.Networking.MachineCIDR` | Immutable after creation |
| `host_prefix` | `HostedCluster.Networking.HostPrefix` | Immutable after creation |
| `aws_additional_compute_security_group_ids` | `HostedCluster.Platform.AWS.SecurityGroups` | Via `Platform` which is `write-mode=mutable` |
| `aws_additional_allowed_principals` | `HostedCluster.Platform.AWS.AdditionalAllowedPrincipals` | Via `Platform` which is `write-mode=mutable` |

#### Fields that CANNOT be added — `service-set` in passthrough

| HCP attribute | Passthrough field | Write-mode |
|---|---|---|
| `fips` | `HostedCluster.FIPS` | service-set |
| `channel` / `channel_group` | `HostedCluster.Channel` | service-set |
| `etcd_encryption` / `kms_key_arn` / `etcd_kms_key_arn` | `HostedCluster.Etcd` / `SecretEncryption` | service-set |
| `auto_node` | `HostedCluster.AutoNode` | service-set |
| `no_cni` | `HostedCluster.Networking` sub-field | service-set |

#### Fields that CANNOT be added — OCM-only concepts

| HCP attribute | Reason |
|---|---|
| `sts` | OCM auth model; Hyperfleet uses ambient AWS credentials |
| `domain_prefix` / `base_dns_domain` / `domain` | OCM DNS management |
| `version` / `current_version` / `upgrade_acknowledgements_for` | OCM-driven upgrade pipeline |
| `autoscaling_enabled` / `replicas` / `compute_machine_type` / `min_replicas` / `max_replicas` / `worker_disk_size` / `ec2_metadata_http_tokens` | Default machine pool concept — Hyperfleet uses explicit nodepools |
| `aws_account_id` / `aws_billing_account_id` | Comes from provider config |
| `audit_log_arn` | OCM-only |
| `shared_vpc` | OCM-specific networking model |
| `external_auth_providers_enabled` | OCM-only |
| `create_admin_user` / `admin_credentials` | OCM-only |
| `registry_config` | OCM-managed |
| `log_forwarders_at_cluster_creation` / `log_forwarder_ids` | OCM-only |

#### Hyperfleet-only fields (no HCP equivalent)

`operator_roles_prefix`, `vpc_id`, `expiration_timestamp`, `phase` (vs HCP's `state`), `management_cluster`, `aws_partition`, `creator_arn`

---

### `rhcs_nodepool_hyperfleet` vs `rhcs_hcp_machine_pool`

#### Fields that CAN be added (mutable in passthrough)

| HCP attribute | Passthrough location | Notes |
|---|---|---|
| `aws_node_pool.additional_security_group_ids` | `Platform.AWS.SecurityGroups` | List of SG IDs |
| `aws_node_pool.capacity_reservation_id` | `Platform.AWS.Placement.CapacityReservation.ID` | |
| `aws_node_pool.capacity_reservation_preference` | `Platform.AWS.Placement.CapacityReservation.Preference` | Enum: `None` / `Open` / `CapacityReservationsOnly` (HyperShift values, not OCM values) |
| `aws_node_pool.image_type` | `Platform.AWS.ImageType` | Enum: `Linux` / `Windows` |

#### Hyperfleet-only fields (not in HCP machine pool — from `AWSNodePoolPlatform.Placement`)

| Attribute | Passthrough location | Notes |
|---|---|---|
| `aws_node_pool.market_type` | `Platform.AWS.Placement.MarketType` | Enum: `OnDemand` / `Spot` / `CapacityBlocks` |
| `aws_node_pool.tenancy` | `Platform.AWS.Placement.Tenancy` | AWS tenancy string |
| `aws_node_pool.spot_max_price` | `Platform.AWS.Placement.Spot.MaxPrice` | Only valid when `market_type=Spot` |

#### Fields that CANNOT be added — `service-set` in passthrough

| HCP attribute | Passthrough field | Write-mode |
|---|---|---|
| `autoscaling` (min/max/enabled) | `NodePoolSpecPassthrough.AutoScaling` | service-set |
| `taints` | `NodePoolSpecPassthrough.Taints` | service-set |
| `tuning_configs` | `NodePoolSpecPassthrough.TuningConfig` | service-set |
| `aws_node_pool.node_drain_grace_period` | `NodePoolSpecPassthrough.NodeDrainTimeout` | service-set |

#### Fields that CANNOT be added — OCM-only

| HCP attribute | Reason |
|---|---|
| `kubelet_configs` | OCM-only |
| `version` / `current_version` / `upgrade_acknowledgements_for` | OCM-driven upgrades |
| `aws_node_pool.ec2_metadata_http_tokens` | OCM default machine pool concept |
| `status` (current_replicas, message) | OCM-specific status fields |
| `availability_zone` (computed) | Derived from subnet in OCM; not tracked in Hyperfleet passthrough |

## Tests

### What exists

| Layer | Location | Coverage |
|---|---|---|
| Unit (cluster) | `provider/hyperfleet/resource_test.go` | `populateState` field mapping, `computeRolesRef` / `prefixAndPartitionFromRolesRef` round-trips |
| Unit (nodepool) | `provider/hyperfleet/nodepool_resource_test.go` | `buildNodePoolSpec` (basic fields, disk size, labels, tags), `populateNodePoolState` |
| Subsystem | `subsystem/hyperfleet/cluster_resource_test.go` | Provider-config diagnostics: missing `hyperfleet_url`, missing `aws_account_id`, `aws_region` mismatch |
| E2E sanity | `tests/e2e/hyperfleet/sanity_test.go` | Full create/read/destroy cycle: VPC → cluster (wait Ready) → IAM → two nodepools (wait Ready) → replica scale (np2: 1→2) → LIFO teardown |

Run the E2E suite:
```
HYPERFLEET_URL=https://<platform-api-host> make e2e-hyperfleet
```

Optional overrides (all default to derived values):
```
HYPERFLEET_CLUSTER_NAME   # default: hf-e2e-<unix-timestamp>  (max 18 chars)
HYPERFLEET_OPERATOR_ROLES_PREFIX
HYPERFLEET_VPC_CIDR       # default: 10.0.0.0/16
HYPERFLEET_NODEPOOL_INSTANCE_TYPE  # default: m5.xlarge
```

AWS credentials must be set in the environment (standard `AWS_*` vars, shared config, or instance profile). The test derives the AWS region from the host portion of `HYPERFLEET_URL`.

### Test gaps

The following scenarios are not covered today:

| Gap | Priority | Notes |
|---|---|---|
| Subsystem CRUD coverage | High | The `subsystem/hyperfleet/` suite covers provider-config diagnostics only. Full CRUD subsystem coverage still needs a Hyperfleet API stub (fake Kubernetes API server or interface mock). |
| Cluster `expiration_timestamp` update | Medium | The only mutable cluster field besides future additions; currently only exercised implicitly. |
| Nodepool `labels` update | Medium | Mutable but not exercised via apply after creation. |
| Nodepool `auto_repair` update | Medium | Mutable but not covered in e2e. |
| Nodepool import | Medium | `ImportState` is implemented but not tested e2e. |
| Cluster import | Medium | `ImportState` is implemented but not tested e2e. |
| `ignore_deletion_error=true` path | Low | Tests the nodepool delete error suppression; can be unit-tested with a mock. |
| Multiple availability zones | Low | E2E only exercises single-AZ clusters and nodepools. |
| Teardown when cluster never reached Ready | Low | Cluster deletion and IAM/VPC cleanup when cluster is in Provisioning phase. |
| Concurrent nodepool creates | Low | Parallel Ginkgo nodes are not exercised. |

### E2E teardown ordering

DeferCleanup blocks register in LIFO order. The intended destruction sequence is:

```
np2 destroy → np1 destroy → cluster destroy (+ WaitUntil deleted) → IAM destroy → VPC destroy (with retry)
```

The cluster `WaitUntil` call blocks until the cluster object returns 404, guaranteeing that IAM and VPC teardown does not race with the operator releasing AWS resources. The VPC destroy is wrapped in `Eventually` (30-minute timeout, 2-minute polling) to tolerate asynchronous AWS resource release by the cluster operator.

### Known quirks

- **`auto_repair` is not read back from the API.** The server populates `spec.nodePool.management.autoRepair` (an internal HyperShift path) rather than `spec.autoRepair` (the public API field). `populateNodePoolState` preserves the caller-supplied value to avoid a false inconsistency. Fix pending on server-side.
- **Nodepool phase after np1 delete.** np1 is deleted with `ignore_deletion_error=false` but its phase will stay `Deleting` due to a PDB. The e2e test does a best-effort `WaitUntil` with a 1-minute timeout and logs the observed phase rather than asserting a specific outcome.
- **VPC destroy async release.** The Hyperfleet operator continues releasing ENIs and security groups after the cluster Kubernetes object is deleted (404). VPC terraform destroy will fail if run immediately after cluster deletion. The 2-minute polling retry in the e2e teardown absorbs this window.
