# Resource schema

Router: [`resource-overview.md`](resource-overview.md). Attribute names: [`naming.md`](../naming.md). Shared helpers: [`package-common.md`](../packages/package-common.md). Secrets: [`security.md`](../security.md). Contract changes: [`breaking-changes.md`](../breaking-changes.md). State structs: [`resource-model.md`](resource-model.md). API/apply diagnostics (not validators): [`errors.md`](../errors.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN deciding what the user must set versus what the API returns → [Required, Optional, Computed](#required-optional-computed)
- WHEN checking config before any API call (conflicts, enums, cross-field rules) → [Validation (no client)](#validation-no-client)
- WHEN the product rule is “if omitted, use this value” → [Defaults](#defaults)
- WHEN a plan modifier is needed (immutability / replace, or carry forward a stable API value) → [Plan modifiers](#plan-modifiers)
- WHEN the API returns a nested object or list/set of objects (optional parent OK; fields inside required when present) — e.g. component route → hostname + TLS → [Nested attributes and collections](#nested-attributes-and-collections)
- DEFAULT: WHEN adding or reshaping attributes, open [Schema method](#schema-method) first, then any matching WHEN above. If nothing matches, Schema method alone is enough.

## Schema method

WHEN implementing or changing a resource schema:
- MUST: Name attributes and blocks per [`naming.md`](../naming.md).
- MUST: Prefer attribute semantics per [`architecture.md`](../architecture.md) (OCM/ROSA / CLI parity) when the provider can express the same contract.
- DEFAULT: WHEN editing an existing package, match that package.
- EXAMPLE (lean Schema shape — attributes, validators, plan modifiers, defaults): `provider/imagemirror`. Prefer a package closest to your feature over copying unrelated CRUD from that package.

NOTE: WHEN deprecating an attribute or block before removal — follow [`breaking-changes.md`](../breaking-changes.md#schema-deprecations); do not invent a parallel deprecation process here.

## Required, Optional, Computed

WHEN deciding what practitioners configure versus what the provider/API owns:
- MUST: Prefer user-set versus server-set semantics per [`architecture.md`](../architecture.md).
- WHEN changing Required ↔ Optional ↔ Computed (or Default) on an existing attribute: MUST follow [`breaking-changes.md`](../breaking-changes.md).
- DEFAULT: Framework attribute behaviors — https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes
- EXAMPLE: `provider/imagemirror` (`cluster_id` Required; `type` Optional+Computed+Default; `id` Computed).

WHEN adding an Optional boolean that enables an API feature where the API omits the feature object entirely when disabled (Read sets the attribute to null):
- MUST NOT: Leave it Optional-only if `false` and `null` produce different Terraform behavior. Terraform treats explicit `false` and null as distinct values — if Read always returns null when the feature is absent, a user who writes `= false` gets a perpetual plan diff (`false` → `null` on every refresh).
- MUST: Use one of: (a) `IgnoreFalse()` plan modifier to normalize `false` → `null` in the plan so `false` is never stored in state (preferred — no schema change, no drift), (b) Optional+Computed so Read can set `false` explicitly when the API omits the feature, or (c) keep it Optional-only and accept null as the only way to express "not enabled" — but then document in the Description that `= false` is not a supported value.
- EXAMPLE: `provider/machinepool/hcp/aws_node_pool.go` `use_spot_instances` — uses `IgnoreFalse()` before `ImmutableBool()` so that `= false` is treated as "not set".

## Validation (no client)

WHEN validating configuration without calling OCM (or other APIs):
- MUST NOT: Rely on resource `Configure` (or a live API client) inside attribute validators, `ConfigValidators`, or `ValidateConfig` — those run before provider Configure may have completed.
- MUST: Put client-dependent checks in Create/Read/Update/Delete (or other RPCs that run with a configured provider).
- NOTE: Schema-level validators and `ConfigValidators` run automatically on all operations. In-method checks inside Create/Update do not — they must be mirrored manually. See [`resource-create.md`](resource-create.md) and [`resource-update.md`](resource-update.md).
- EXAMPLE (`ConfigValidators`): `provider/machinepool/classic` — copy only the validation pattern you need. Reuse helpers per [`package-common.md`](../packages/package-common.md).

## Defaults

WHEN the product rule is “if omitted, use this value”:
- MUST: Add `Default` only on resource attributes (not provider or data source).
- MUST NOT: Add a default that conflicts with the API default unless documented — see [`breaking-changes.md`](../breaking-changes.md).
- EXAMPLE: `provider/imagemirror` (`stringdefault`) or `provider/machinepool` (`booldefault`).

## Plan modifiers

Common Framework modifiers: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#common-use-case-attribute-plan-modifiers

WHEN the product/API rule is “this value cannot change in place” (e.g. name immutable → replace):
- MUST: Use `RequiresReplace()` (typed package, e.g. `stringplanmodifier`).
- EXAMPLE: `provider/imagemirror` (`cluster_id`, `source`).

WHEN the value comes from the API and should remain stable across apply cycles unless the API changes it, and copying prior non-null state into the plan is correct:
- MUST: Prefer `UseStateForUnknown()`.
- EXAMPLE: `provider/imagemirror` (`id`); many attrs under `provider/clusterrosa` and machine pools.

WHEN a nested optional+computed child may stay null in config while the API returns a value, and `UseStateForUnknown` would preserve null and risk plan/apply inconsistency (framework ≥1.15.1 — https://github.com/hashicorp/terraform-plugin-framework/issues/1197):
- MUST: Prefer `UseNonNullStateForUnknown()` (framework ≥1.17).
- EXAMPLE: `provider/registry_config` (`platform_allowlist_id`).

WHEN writing attribute plan modifiers under list/set nested attributes or blocks:
- MUST NOT: Assume prior-state elements stay aligned by index after reorder/remove.

WHEN replace behavior or plan semantics change for existing configurations:
- MUST: Follow [`breaking-changes.md`](../breaking-changes.md).

NOTE: Custom modifiers — search `provider/common/planmodifiers` (and feature `common`) before adding; e.g. `ImmutableString()` used from `provider/machinepool/hcp/aws_node_pool`. See [`package-common.md`](../packages/package-common.md).

## Nested attributes and collections

WHEN the API exposes a nested structure (object or list/set of objects — e.g. component route with hostname + TLS fields):
- MUST: Prefer nested attributes over blocks for new work unless an analogous package already uses blocks for the same shape.

WHEN the parent object or list is optional, but fields inside are mandatory when that parent/element is present:
- MUST: Treat Required nested fields as required only when that element/object is present. Omitting the optional parent is valid; including the parent (or a list element) without required children is a validation error.
- MUST: Make that clear in attribute Descriptions (users often confuse “optional list” with “optional fields inside each element”).
- EXAMPLE: `provider/machinepool/classic` `taints` — list Optional; each object’s `key` / `value` / `schedule_type` Required.

WHEN choosing list versus set versus map for a collection from the API:
- MUST: Use list when order is part of the contract; set when order does not matter and duplicates are not meaningful; map when the API is keyed.
- MUST NOT: Pick list only for convenience if the API treats the collection as unordered — that invites spurious diffs when the API reorders.
- NOTE: Suppressing reorder-only drift when writing state → [`resource-read.md`](resource-read.md) / [`resource-update.md`](resource-update.md).

WHEN an API-keyed collection loses keys or nested objects are omitted from config:
- MUST: See [`resource-update.md`](resource-update.md) (clear-on-shrink).

NOTE: Populating nested structures from the API into state — [`resource-read.md`](resource-read.md) (and create/update).
