# Resource create

Router: [`resource-overview.md`](resource-overview.md). Model and collection nulls: [`resource-model.md`](resource-model.md). Schema: [`resource-schema.md`](resource-schema.md). Diagnostics and API failures: [`errors.md`](../errors.md). Logging: [`logging.md`](../logging.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN implementing Create or mapping plan → API create → state → [Create method](#create-method)
- WHEN this resource’s Create must wait until a parent cluster is ready → [Wait for parent cluster](#wait-for-parent-cluster)
- WHEN adding or changing a Day 2 capability (where it lives: cluster attribute versus its own resource) → [Day 2 capabilities](#day-2-capabilities)
- DEFAULT: WHEN adding or changing create behavior, open [Create method](#create-method) first, then any matching WHEN above.

## Create method

WHEN implementing Create:
- MUST: Read input from `CreateRequest.Plan` (not Config) so plan modifiers and defaults are reflected.
- MUST: Prefer OCM/ROSA create semantics per [`architecture.md`](../architecture.md) when the provider can express the same behavior.
- MUST: Write only known or null values to `CreateResponse.State` — never unknown. Copy every null or known plan value into response state as-is; only unknown plan values may be filled from the API response.
- MUST: Follow [`resource-model.md`](resource-model.md) for List/Set/Map nulls and [`resource-schema.md`](resource-schema.md) nested Required rules when populating state from the API.
- MUST: Handle create failures (including duplicate remote object → import) per [`errors.md`](../errors.md).
- DEFAULT: WHEN editing an existing package, match that package. For new types, follow HashiCorp Create (https://developer.hashicorp.com/terraform/plugin/framework/resources/create) where it does not conflict with this repo’s rules.
- EXAMPLE: `provider/imagemirror` Create (plan get → validate cluster ready/HCP → API add → set id/timestamps → `State.Set`). Prefer a package closest to your feature.

## Wait for parent cluster

This section is **not** about designing Day 2 product APIs. It is only: Create of **this** resource needs the cluster (or parent) already ready.

WHEN Create (or a pre-check) requires an existing cluster to be ready before this resource’s API call:
- MUST: Follow the nearest analogous package — typically `clusterWait.WaitForClusterToBeReady`, or an explicit ready-state check.
- MUST: Prefer readiness and validation messages per [`architecture.md`](../architecture.md) (ROSA CLI parity) when the provider can express them.
- DEFAULT: Reuse existing wait helpers; do not invent a new wait pattern.
- EXAMPLE: `provider/tuningconfigs`, `provider/machinepool/classic`, `provider/machinepool/hcp`, `provider/logforwarder`, `provider/kubeletconfig` — `WaitForClusterToBeReady` then create.
- EXAMPLE: `provider/imagemirror` — get cluster and require ready / HCP before create (same idea, explicit checks).

WHEN implementing **cluster** create itself and the user opts into waiting until the cluster is ready:
- EXAMPLE: `wait_for_create_complete` on `provider/clusterrosa/hcp` (and classic) — that is cluster Day 1 create polling, not a child-resource wait.

NOTE: Timeouts — WHEN the resource needs configurable create/update/delete duration, use `terraform-plugin-framework-timeouts` like the nearest analogous package (prefer nested attributes for new work; blocks only when migrating an existing block-based timeout).

## Day 2 capabilities

**Day 2** (code also says **post-create**) means a capability that applies **after** the cluster already exists. `WaitForClusterToBeReady` inside a child resource’s own Create is expected and is **not** what this section constrains. Embedding a new Day 2 API call inside cluster Create/Update is what this section constrains.

WHEN adding a capability that is Day 2:
- DEFAULT: Prefer a **new resource** that takes the cluster id so Terraform can express dependencies (create order, destroy order), over a new attribute on the cluster resource that runs extra API calls inside cluster Create/Update.
- MUST NOT: For **new** Day 2 behavior embedded in cluster Create/Update, succeed cluster creation while leaving that capability unset or only partially applied.
- NOTE: When the capability is modeled as a separate resource, cluster creation may complete before that resource is applied.
- EXAMPLE (preferred shape): `provider/imagemirror`, `provider/logforwarder`, `provider/machinepool/classic`, `provider/machinepool/hcp` — own resource, `cluster` / `cluster_id`, wait-for-ready in Create, then that resource’s API.
- EXAMPLE (legacy embed — do not copy for new work): `auto_node` on `provider/clusterrosa/hcp` — Day 2 PATCH after cluster ready inside cluster Create; requires `wait_for_create_complete`.

WHEN editing an existing cluster attribute that already embeds Day 2 behavior:
- DEFAULT: Match that package, including warning-on-failure if that is the established pattern. Severity and safety-control failures: [`errors.md`](../errors.md).

NOTE: Splitting an existing embedded Day 2 attribute into its own resource is a user-visible contract change — see [`breaking-changes.md`](../breaking-changes.md).
