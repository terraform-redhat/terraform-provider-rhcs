# Resource import

Router: [`resource-overview.md`](resource-overview.md). Read after import: [`resource-read.md`](resource-read.md). Diagnostics: [`errors.md`](../errors.md). Docs: [`docs-and-examples.md`](../docs-and-examples.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN implementing ImportState or deciding what import sets in state → [Import method](#import-method)
- WHEN the import identifier is more than a single id (e.g. cluster + resource) → [Composite import ID](#composite-import-id)
- WHEN import is not appropriate for this resource → [Omit import](#omit-import)
- DEFAULT: WHEN adding or changing import behavior, open [Import method](#import-method) first, then any matching WHEN above. If unsure whether import applies, open [Omit import](#omit-import) first.

## Import method

WHEN implementing ImportState:
- MUST: Keep ImportState thin — parse the import id, set the identity attributes needed for Read, and let Read populate the rest (see [`resource-read.md`](resource-read.md)).
- MUST: Validate the import id format and add an **error** diagnostic on invalid format (then return) — see [`errors.md`](../errors.md) for lookup failures in Read.
- MUST: Document the import identifier format for practitioners (resource docs / examples) when import is supported — see [`docs-and-examples.md`](../docs-and-examples.md).
- DEFAULT: For new types, follow HashiCorp Import (https://developer.hashicorp.com/terraform/plugin/framework/resources/import) where it does not conflict with this repo’s rules. When editing an existing package, match that package (including delimiter and attribute names).
- EXAMPLE: `provider/imagemirror` ImportState — `cluster_id:image_mirror_id` → `SetAttribute` on `cluster_id` and `id`.
- EXAMPLE (passthrough id): `provider/defaultingress` — `ImportStatePassthroughID` into `cluster`.

NOTE: Design the import identifier so practitioners can form it from values they already have (OCM/API id, cluster id, CLI). Prefer stable remote ids over ephemeral or display-only names.

NOTE: Changing an established import id format is a user-visible contract change — see [`breaking-changes.md`](../breaking-changes.md).

## Composite import ID

WHEN the resource is scoped to a parent (usually a cluster) and import needs more than one identifier:
- MUST: Use a single import string that encodes every id Read needs; validate part count and non-empty parts before `SetAttribute`.
- MUST: Match the nearest analogous package for delimiter and order (this repo uses both `,` and `:` — do not invent a third style for a sibling of an existing type).
- DEFAULT: Prefer ROSA CLI / Classic–HCP sibling import messaging when the provider can express the same identifier shape.
- EXAMPLE: `provider/machinepool/classic`, `provider/logforwarder`, `provider/tuningconfigs` — `<cluster_id>,<resource_id>`.
- EXAMPLE: `provider/imagemirror` — `cluster_id:image_mirror_id`.

## Omit import

WHEN the resource is not a durable remote object practitioners re-adopt with `terraform import` (e.g. wait/poll helper, local/input-only helper with nothing meaningful to import):
- MUST: Omit `ImportState` so Terraform reports import not implemented — do not add an empty or no-op `ImportState`.
- DEFAULT: Prefer omitting import over a stub implementation; document in resource docs if practitioners might expect import.
- EXAMPLE: `provider/clusterwaiter` (`rhcs_cluster_wait`) — no `ImportState` (wait helper; Read is also a no-op).
- EXAMPLE (do not copy): `provider/oidcconfiginput` — empty `ImportState` (“Do Nothing”); for new types, omit the method instead.
