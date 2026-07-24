# Resource read

Router: [`resource-overview.md`](resource-overview.md). Model and collection nulls: [`resource-model.md`](resource-model.md). Nested Required / list-set order: [`resource-schema.md`](resource-schema.md). Diagnostics: [`errors.md`](../errors.md). Import: [`resource-import.md`](resource-import.md). Logging: [`logging.md`](../logging.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN implementing Read or refreshing state from the API → [Read method](#read-method)
- WHEN mapping API nested objects or collections into state → [Populate state from API](#populate-state-from-api)
- DEFAULT: WHEN adding or changing read behavior, open [Read method](#read-method) first, then any matching WHEN above.

## Read method

WHEN implementing Read:
- MUST: Read current values from `ReadRequest.State`, then refresh refreshable attributes into `ReadResponse.State` from the API.
- MUST: Refresh all refreshable attributes from the API so Terraform can show drift and import stays thin.
- MUST: Prefer read/refresh semantics per [`architecture.md`](../architecture.md) when the provider can express the same behavior.
- MUST: Follow [`resource-model.md`](resource-model.md) for List/Set/Map typed nulls when writing collections into state.
- MUST: Handle not-found and other refresh failures per [`errors.md`](../errors.md).
- DEFAULT: For new types, follow HashiCorp Read (https://developer.hashicorp.com/terraform/plugin/framework/resources/read) where it does not conflict with this repo’s rules. When editing an existing package, match that package (even when it differs from Framework recommendations).
- EXAMPLE: `provider/imagemirror` Read (state get → API get → update fields → `State.Set`; 404 → `RemoveResource`). Prefer a package closest to your feature.

NOTE: Read also runs after import — see [`resource-import.md`](resource-import.md).

## Populate state from API

WHEN writing nested objects from the API into state:
- MUST: Set a nested object/block only when every Required attribute can be set from the API; otherwise leave it unset (`nil` / `ObjectNull`) — do not write a partial object with null Required children.
- MUST: When the API omits or clears a value, assign the matching typed null (or the empty value that matches the attribute contract) so refresh does not keep stale data — see [`resource-model.md`](resource-model.md).
- NOTE: Leaving the nested object unset while config still has it can show plan drift. If the API never returns a Required field, fix the schema contract (e.g. write-only, Optional+Computed, plan modifiers) — match the analogous package; do not invent partial nested state.
- EXAMPLE: `provider/clusterrosa` `FlattenAdminCredentials` → `ObjectNull` when empty; `provider/logforwarder` null nested S3/CloudWatch when absent.

WHEN the refreshed value is semantically equal to prior state (for example JSON key order or whitespace-only differences):
- MUST: Prefer preserving the prior state value to avoid spurious drift, matching the nearest analogous package.

WHEN the API returns an unordered collection in a different order than state:
- NOTE: Choosing list versus set versus map is owned by [`resource-schema.md`](resource-schema.md). Suppressing reorder-only drift when writing state belongs here or in Update — follow the analogous package.
