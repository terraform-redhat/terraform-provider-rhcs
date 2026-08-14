# Resource update

Router: [`resource-overview.md`](resource-overview.md). Schema / plan modifiers: [`resource-schema.md`](resource-schema.md). Model / nulls: [`resource-model.md`](resource-model.md). Diagnostics: [`errors.md`](../errors.md). Logging: [`logging.md`](../logging.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN implementing Update or mapping plan → API update → state → [Update method](#update-method)
- WHEN the API only accepts changed fields or per-attribute PATCHes → [Partial updates](#partial-updates)
- WHEN config removes keys from a map/object (shrink or omit), or a managed map must reconcile to full desired state (including create) → [Clear on shrink](#clear-on-shrink)
- DEFAULT: WHEN adding or changing update behavior, open [Update method](#update-method) first, then any matching WHEN above.

## Update method

WHEN implementing Update:
- MUST: Read input from `UpdateRequest.Plan` (not Config) so plan modifiers and defaults are reflected. Use prior `UpdateRequest.State` when comparing for changes.
- MUST: Save every null or known plan value into response state exactly as-is; only unknown plan values may be filled from the API.
- MUST: Follow [`resource-model.md`](resource-model.md) and [`resource-read.md`](resource-read.md) populate rules when writing nested objects/collections into state after update.
- MUST: Handle not-found and other update failures per [`errors.md`](../errors.md) (no `RemoveResource` in Update).
- MUST: Mirror in-method validations from Create — schema-level validators ([`resource-schema.md`](resource-schema.md) `Validators`) run automatically on all operations, but checks inside `Create()` (cross-field rules, region matching, mutual exclusion) must be replicated in `Update()` for any attribute that can change after creation.
- WHEN the resource does not support in-place update: MUST leave Update empty (or unused) and ensure configurable attributes that cannot change use `RequiresReplace()` — see [`resource-schema.md`](resource-schema.md).
- DEFAULT: For new types, follow HashiCorp Update (https://developer.hashicorp.com/terraform/plugin/framework/resources/update) where it does not conflict with this repo’s rules. When editing an existing package, match that package.
- EXAMPLE: `provider/imagemirror` Update (plan + state get → PATCH type/mirrors → keep id → `State.Set`). Prefer a package closest to your feature.

NOTE: Stable computed attributes that must not appear as `(known after apply)` after update — declare `UseStateForUnknown()` (or `UseNonNullStateForUnknown` when required) in schema — see [`resource-schema.md`](resource-schema.md).

## Partial updates

WHEN the OCM/API update requires only changed attributes or separate per-attribute calls:
- MUST: Compare plan to prior state to detect changes before building the request.
- MUST: Prefer existing helpers (e.g. `provider/common` `ShouldPatch*`) when they fit — see [`package-common.md`](../packages/package-common.md).
- DEFAULT: Match the nearest analogous package (Classic/HCP siblings) and ROSA CLI edit behavior.
- EXAMPLE: `provider/machinepool/classic` — `ShouldPatchString` / `ShouldPatchInt` / `ShouldPatchMap` (and similar) before PATCH.

## Clear on shrink

WHEN a map or nested object is part of desired state and keys/entries are removed from config (or the whole map is omitted/nulled when absent means clear):
- MUST: Treat Terraform config as **full desired state** for that managed map — do not leave stale remote keys when Terraform no longer manages them (unless the attribute is explicitly documented as patch-merge and that matches CLI/sibling behavior).
- MUST: Update the API so removed entries are cleared. WHEN the product has a fixed set of managed keys, prefer **reset those keys then overlay the plan** (do not depend only on keys present in prior state or a complete first API read).
- MUST: Align null versus empty collections in state with the attribute contract to avoid permadiff — see [`resource-model.md`](resource-model.md) and [`resource-read.md`](resource-read.md).
- DEFAULT: Prefer Classic–HCP sibling behavior for the same field. ROSA CLI partial/merge edit is fine for the CLI; do not copy it into Terraform unless merge-only is intentional and documented — see [`architecture.md`](../architecture.md) (CLI versus desired-state).
- EXAMPLE: `provider/defaultingress/classic` — `component_routes` clear/shrink via `ResetComponentRoutes` when the map changes.

NOTE: On **create**, the first plan may not preview clearing of pre-existing remote keys (no prior state yet). Apply must still reconcile the managed map to config so those keys are cleared in that apply — do not rely on a later update cycle.
