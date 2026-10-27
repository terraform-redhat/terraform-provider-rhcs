# Seaplane resources

This document covers the architecture, file layout, schema design, and test strategy for the Seaplane resources (`rhcs_cluster_seaplane` and `rhcs_nodepool_seaplane`). It also records the deliberate gaps relative to their HCP equivalents and the reasons behind them.

## What is Seaplane

Seaplane exposes a Kubernetes-native Platform API for managing ROSA HCP clusters. It is distinct from OCM in several important ways:

| Concern | OCM (HCP) | Seaplane |
|---|---|---|
| Authentication | OCM offline token | Ambient AWS credentials (env vars, shared config, instance profile) |
| API style | OCM REST SDK | Kubernetes-style clientset (Get/List/Create/Update/Delete/Watch) |
| Resource model | OCM resources (clusters, machine pools) | Kubernetes custom resources (`Cluster`, `NodePool` in `seaplane.io/v1alpha1`) |
| HyperShift fields | Mediated via OCM API model | Passed through to HyperShift types directly (`HostedClusterSpecPassthrough`, `NodePoolSpecPassthrough`) |
| Operator roles | Managed by OCM IAM module | User-managed; constructed from `operator_roles_prefix` and account ID |
| OIDC provider | Created by OCM | Created by the user after cluster creation (IAM tf-manifest) |

Because Seaplane resources share no auth or API surface with OCM, they cannot use the OCM SDK client (`cmv1`) and cannot participate in OCM-level features (channels, machine pool autoscaling, tuning configs, kubelet configs, etc.).

## Architecture decisions

### No OCM token; provider config is different
The provider `rhcs` block for Seaplane resources requires `seaplane_url` and `aws_account_id`. `aws_account_id` is sent as the mandatory `X-Amz-Account-Id` SigV4 signed header. `aws_caller_arn` is optional: when supplied it is added as the `X-Amz-Caller-Arn` signed header (and becomes the cluster `CreatorARN`); when omitted, requests are still signed and authenticated with the ambient AWS credentials and account ID. `aws_region` is also optional — it is derived from the `seaplane_url` execute-api hostname when unset. No `token` or `url` (OCM gateway) is needed. The client is `seaplane.Interface` from `rosa-seaplane-api/clientset`, not `*cmv1.Client`.

### Passthrough write-mode governs what users can set
The Seaplane API annotates every HyperShift field with `+seaplane:write-mode`:
- `mutable` — users can set and update the field.
- `service-set` — the platform operator owns the field; user writes are rejected.

Fields marked `service-set` cannot be exposed as writable resource attributes. See the gap tables below for the full breakdown.

### Subsystem tests use a local Platform API stub
The standard `subsystem/hcp/` and `subsystem/classic/` suites rely on a stubbed OCM HTTP server (`TestServer`). The Seaplane Platform API uses a Kubernetes-style API protocol, so tests under `subsystem/seaplane/` exercise the Terraform provider against a local Seaplane HTTP stub instead. The suite setup in `seaplane_test.go` still provides fake AWS credentials and a local OCM server for provider initialization; feature tests use the subsystem framework's local server for Seaplane requests.

Subsystem tests cover more than provider configuration: they can verify Terraform configuration sent to the Platform API, response-to-state mapping, resource lifecycle behavior, and diagnostics without a live Seaplane environment. Use the existing local-server patterns in `cluster_resource_test.go`, `nodepool_resource_test.go`, and `oidcconfig_resource_test.go`.

### Worker instance profile is computed, not exposed
The Seaplane cluster object carries the operator role ARNs (`RolesRef`) but **not** the worker IAM instance profile. If the nodepool spec omits it, CAPA defaults to a non-existent `<infra-id>-worker-profile` and the create fails with `Invalid IAM Instance Profile name`. The provider therefore computes it during `Create`: it fetches the parent cluster, recovers the operator roles prefix from `RolesRef` (via `prefixAndPartitionFromRolesRef`), and sets `spec.NodePool.Platform.AWS.InstanceProfile = <prefix>-ROSA-Worker-Role` — matching the IAM manifest naming convention (`tests/tf-manifests/aws/iam-roles/seaplane/main.tf`). On `Update` the value is preserved from the live object.

This value is **derived, not user-supplied**, and is intentionally *not* a schema attribute or state field — there is no per-nodepool customization today. If a future use case needs custom instance profiles, promote it to an optional `aws_node_pool.instance_profile` attribute and read it back from the API response in `populateNodePoolState` instead of computing it (see the note in `nodepool_resource.go` `Create`).

### Separate nodepool resource (not the default machine pool)
HCP clusters created via OCM have a default machine pool managed by `rhcs_cluster_rosa_hcp`. Seaplane clusters have no managed default pool — all node pools are explicit `rhcs_nodepool_seaplane` resources. The cluster resource exposes only control-plane configuration.

### Fire-and-forget delete for nodepools
The Seaplane API returns 200 for a nodepool delete request immediately; the operator drains and deprovisions asynchronously. The Terraform resource clears state on a 200 or 404. A PDB on a cluster may block eviction indefinitely — those nodepools are expected to stay in `Deleting` phase until the cluster itself is deleted (which cascades cleanup).

### Cluster deletion is synchronous in the e2e test
The e2e teardown uses `WaitUntil` (a synchronous ticker loop from `bridge_wrappers_generated.go`) to block until the cluster's 404 is confirmed before IAM and VPC teardown runs. This prevents race conditions where the operator is still releasing VPC resources (ENIs, security groups) when `terraform destroy` runs on the VPC.

### VPC destroy uses a retry loop
After the cluster object is deleted (404), the cluster operator continues releasing AWS resources asynchronously. The e2e teardown wraps `vpcSvc.Destroy()` in an `Eventually` with a 30-minute timeout and 2-minute polling interval to tolerate this delay. Classic load balancers and non-default security groups are pruned first using `helper.DeleteClassicLoadBalancers` / `helper.DeleteNonDefaultSecurityGroups` before the retry loop.

### VPC teardown purges the private hosted zone
The cluster operator writes CNAME records (`api.<name>.hypershift.local`, `*.apps.<name>.hypershift.local`) into the private hosted zone created by the VPC manifest (`aws_route53_zone.seaplane`). Route53 refuses to delete a non-empty zone, so `terraform destroy` of the VPC fails with `HostedZoneNotEmpty` while those records remain. The e2e teardown calls `helper.PurgeHostedZoneRecords(region, vpcOut.HostedZoneID)` before `vpcSvc.Destroy()` to delete every record except the NS/SOA pair (which Route53 manages with the zone lifecycle itself). The hosted zone ID is surfaced by the VPC manifest as the `hosted_zone_id` output.

### Cluster name length limit
The Seaplane operator creates a namespace named `cluster-<36-char-uuid>-<cluster-name>`. Kubernetes namespaces are capped at 63 characters. After the fixed prefix and UUID, at most 18 characters remain for the cluster name (`63 - len("cluster-") - 36 - len("-") = 18`). The e2e default name `sp-e2e-<unix-ts>` is within this limit; names beyond 18 characters cause a 422 from the API.

### Import state
Both resources implement `ResourceWithImportState`. The cluster resource imports by cluster UUID. The nodepool resource imports by `<cluster_uuid>/<nodepool_name>`.

## File layout

```
provider/seaplane/
├── resource.go               # rhcs_cluster_seaplane — Schema, CRUD, ImportState
├── state.go                  # ClusterSeaplaneState struct with tfsdk tags
├── resource_test.go          # Unit tests for populateState, computeRolesRef
├── nodepool_resource.go      # rhcs_nodepool_seaplane — Schema, CRUD, ImportState
├── nodepool_state.go         # NodePoolSeaplaneState, NPAWSNodePool structs
└── nodepool_resource_test.go # Unit tests for buildNodePoolSpec, populateNodePoolState

tests/e2e/seaplane/
├── suite_test.go             # Ginkgo suite registration (TestSeaplaneSanity)
└── sanity_test.go            # End-to-end: VPC → cluster → IAM → nodepools → scale → teardown

tests/tf-manifests/rhcs/
├── clusters/seaplane/        # Terraform root for rhcs_cluster_seaplane
│   ├── main.tf               # Provider config + resource block
│   ├── variables.tf          # Input variables
│   └── output.tf             # cluster_id, cluster_name, cluster_phase, cluster_api_url, oidc_issuer
├── node-pools/seaplane/      # Terraform root for rhcs_nodepool_seaplane
│   ├── main.tf
│   ├── variables.tf
│   └── output.tf             # nodepool_id, nodepool_name, nodepool_phase, nodepool_replicas
└── (vpc/, iam/ are under a separate Seaplane path — see exec helpers)

tests/utils/exec/
├── seaplane_cluster.go       # SeaplaneClusterService + SeaplaneClusterArgs/Output
├── seaplane_nodepool.go      # SeaplaneNodePoolService + SeaplaneNodePoolArgs/Output
├── seaplane_iam.go           # SeaplaneIAMService + SeaplaneIAMArgs/Output
└── seaplane_vpc.go           # SeaplaneVPCService + SeaplaneVPCArgs/Output
```

The `manifests.GetSeaplane*ManifestsDir()` helpers in `tests/utils/exec/manifests/manifests.go` map each service to its tf-manifest directory.

## Schema reference

### `rhcs_cluster_seaplane`

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

### `rhcs_nodepool_seaplane`

| Attribute | Type | R/O/C | Immutable | Notes |
|---|---|---|---|---|
| `id` | string | C | — | NodePool UID |
| `name` | string | R | yes | |
| `cluster` | string | R | yes | Parent cluster UUID (`id` output of `rhcs_cluster_seaplane`) |
| `subnet_id` | string | R | yes | AWS subnet for node instances |
| `replicas` | int64 | O | no | Fixed replica count |
| `labels` | map(string) | O | no | Node labels |
| `auto_repair` | bool | R | no | |
| `aws_node_pool.instance_type` | string | R | yes | EC2 instance type |
| `disk_size` | int64 | O | yes | Root volume GiB; changing or removing an explicit value replaces the pool |
| `aws_node_pool.tags` | map(string) | O | no | Additional AWS resource tags |
| `phase` | string | C | — | WaitingForCluster / Provisioning / Ready / Deleting |
| `ignore_deletion_error` | bool | O/C | no | Suppress delete errors; default false |

Import format: `<cluster_uuid>/<nodepool_name>`

## HCP parity gap analysis

### `rhcs_cluster_seaplane` versus `rhcs_cluster_rosa_hcp`

#### Fields that CAN be added (mutable in passthrough)

| HCP attribute | Passthrough location | Notes |
|---|---|---|
| `tags` | `ClusterSpec.Tags` | Immutable after creation; maximum 23 customer tags |
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
| `sts` | OCM auth model; Seaplane uses ambient AWS credentials |
| `domain_prefix` / `base_dns_domain` / `domain` | OCM DNS management |
| `version` / `current_version` / `upgrade_acknowledgements_for` | OCM-driven upgrade pipeline |
| `autoscaling_enabled` / `replicas` / `compute_machine_type` / `min_replicas` / `max_replicas` / `worker_disk_size` / `ec2_metadata_http_tokens` | Default machine pool concept — Seaplane uses explicit nodepools |
| `aws_account_id` / `aws_billing_account_id` | Comes from provider config |
| `audit_log_arn` | OCM-only |
| `shared_vpc` | OCM-specific networking model |
| `external_auth_providers_enabled` | OCM-only |
| `create_admin_user` / `admin_credentials` | OCM-only |
| `registry_config` | OCM-managed |
| `log_forwarders_at_cluster_creation` / `log_forwarder_ids` | OCM-only |

#### Seaplane-only fields (no HCP equivalent)

`operator_roles_prefix`, `vpc_id`, `expiration_timestamp`, `phase` (versus HCP's `state`), `management_cluster`, `aws_partition`, `creator_arn`

---

### `rhcs_nodepool_seaplane` versus `rhcs_hcp_machine_pool`

#### Fields that CAN be added (mutable in passthrough)

| HCP attribute | Passthrough location | Notes |
|---|---|---|
| `aws_node_pool.additional_security_group_ids` | `Platform.AWS.SecurityGroups` | List of SG IDs |
| `aws_node_pool.capacity_reservation_id` | `Platform.AWS.Placement.CapacityReservation.ID` | |
| `aws_node_pool.capacity_reservation_preference` | `Platform.AWS.Placement.CapacityReservation.Preference` | Enum: `None` / `Open` / `CapacityReservationsOnly` (HyperShift values, not OCM values) |
| `aws_node_pool.image_type` | `Platform.AWS.ImageType` | Enum: `Linux` / `Windows` |

#### Seaplane-only fields (not in HCP machine pool — from `AWSNodePoolPlatform.Placement`)

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
| `availability_zone` (computed) | Derived from subnet in OCM; not tracked in Seaplane passthrough |

## Tests

### Required test layers for new Seaplane features

Every new feature onboarded for Seaplane clusters MUST include meaningful **unit, subsystem, and E2E coverage**. Add or extend tests at all three layers for feature behavior and applicable updates; existing tests count only when they exercise and assert the new behavior.

| Layer | Required location | What to verify |
|---|---|---|
| Unit | `provider/seaplane/*_test.go` | Schema and plan behavior, validation, Terraform-to-Platform API expansion, Platform API-to-state flattening, state mapping, helper behavior, and diagnostics as applicable. Include valid and invalid inputs and null/unknown/default cases when those affect the behavior. |
| Subsystem | `subsystem/seaplane/*_test.go` | Run Terraform against the local Seaplane API stub. Assert request method/path and relevant payload fields, returned state, updates/deletes, and error diagnostics for the feature as applicable. Use stubbed API responses; unit tests alone do not verify the provider's Terraform/API integration. |
| E2E | Shared ROSA HCP test or, when permitted below, `tests/e2e/seaplane/` | Exercise the feature against a Seaplane cluster selected by an E2E profile. Assert the user-visible result in the cluster/provider state and clean up resources. The test must be enabled for Seaplane profiles and actually reach the Seaplane backend. |

#### E2E placement: HCP test first

Before creating a Seaplane-specific E2E test, search for an existing ROSA HCP E2E test that covers the same capability.

| Matching ROSA HCP E2E test | Required action |
|---|---|
| Exists | First extend or refactor that test into a shared ROSA HCP/Seaplane workflow, following the shared workflow pattern in [`tests/e2e/hcp_machine_pool_test.go`](../tests/e2e/hcp_machine_pool_test.go). Put backend-neutral operations behind the adapters in `tests/utils/clusterworkflow/` and the lifecycle interfaces in `tests/utils/exec/`; use the Seaplane profile marker/gates such as `ci.SeaplaneValidated`. Keep HCP coverage and add backend-specific assertions only where behavior genuinely differs. Do not add a duplicate Seaplane-only test for the same capability. |
| Does not exist | **If and only if** no matching ROSA HCP E2E test exists, add a dedicated test under `tests/e2e/seaplane/` (package `seaplane`). Follow that package's `suite_test.go`, use the Seaplane E2E profile and shared workflow helpers where applicable, and register cleanup for created resources. |

This decision is part of the feature's E2E coverage requirement: a matching HCP test is the place to introduce shared Seaplane coverage before considering a separate Seaplane test.

### Existing test coverage

| Layer | Location | Coverage |
|---|---|---|
| Unit (cluster) | `provider/seaplane/cluster_resource_test.go` | Cluster schema, tag validation, role-reference mapping, network configuration, and plan behavior |
| Unit (nodepool) | `provider/seaplane/nodepool_resource_test.go` | Nodepool schema, validation, expansion, and state mapping |
| Subsystem (cluster) | `subsystem/seaplane/cluster_resource_test.go` | Provider diagnostics and stubbed Platform API create/read/update/delete flows, including proxy and tag behavior |
| Subsystem (nodepool/OIDC) | `subsystem/seaplane/nodepool_resource_test.go`, `oidcconfig_resource_test.go` | Stubbed nodepool lifecycle and updates, OIDC config creation, and validation diagnostics |
| E2E shared workflow | `tests/e2e/hcp_machine_pool_test.go` | Backend-neutral nodepool workflows selected for ROSA HCP or Seaplane profiles, including scaling, labels/taints/auto-repair, IMDSv2, autoscaling, and tags |
| E2E Seaplane-specific | `tests/e2e/seaplane/` | Seaplane cluster create/destroy sanity, proxy behavior, and scheduler-profile configuration |

Run the E2E suite:
```
SEAPLANE_URL=https://<platform-api-host> make e2e-seaplane
```

Optional overrides (all default to derived values):
```
SEAPLANE_CLUSTER_NAME   # default: sp-e2e-<unix-timestamp>  (max 18 chars)
SEAPLANE_OPERATOR_ROLES_PREFIX
SEAPLANE_VPC_CIDR       # default: 10.0.0.0/16
SEAPLANE_NODEPOOL_INSTANCE_TYPE  # default: m5.xlarge
```

AWS credentials must be set in the environment (standard `AWS_*` vars, shared config, or instance profile). The test derives the AWS region from the host portion of `SEAPLANE_URL`.

### E2E gaps

The following live-cluster scenarios are not covered today; unit and subsystem coverage may exist independently:

| Gap | Priority | Notes |
|---|---|---|
| Cluster `expiration_timestamp` update | Medium | The subsystem suite covers the update through the local API stub; no Seaplane live-cluster E2E update exists. |
| Nodepool `labels` update | Medium | The subsystem suite covers the update; no Seaplane live-cluster E2E update exists. |
| Nodepool `auto_repair` update | Medium | The subsystem suite covers the update; no Seaplane live-cluster E2E update exists. |
| Nodepool import | Medium | `ImportState` is implemented but not tested e2e. |
| Cluster import | Medium | `ImportState` is implemented but not tested e2e. |
| `ignore_deletion_error=true` path | Low | Nodepool delete-error suppression is not exercised against a live Seaplane environment. |
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
- **VPC destroy async release.** The Seaplane operator continues releasing ENIs and security groups after the cluster Kubernetes object is deleted (404). VPC terraform destroy will fail if run immediately after cluster deletion. The 2-minute polling retry in the e2e teardown absorbs this window.
