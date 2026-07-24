# Resource model

Router: [`resource-overview.md`](resource-overview.md). Schema contract: [`resource-schema.md`](resource-schema.md). File layout (`*_state.go`): [`package-layout.md`](../packages/package-layout.md). Attribute names: [`naming.md`](../naming.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN defining or changing the Go structs used with Plan/State/Config → [State structs](#state-structs)
- WHEN a List, Set, or Map field must round-trip through Terraform without type errors → [Collections and nulls](#collections-and-nulls)
- DEFAULT: WHEN adding or reshaping the model, open [State structs](#state-structs) first, then any matching WHEN above. If nothing matches, State structs alone is enough.

## State structs

WHEN implementing or changing the resource model:
- MUST: `tfsdk` tags must match the schema attribute names defined in [`resource-schema.md`](resource-schema.md) and [`naming.md`](../naming.md).
- MUST: Prefer field semantics per [`architecture.md`](../architecture.md) when mapping API shapes into the model.
- DEFAULT: WHEN editing an existing package, match that package’s model shape. Prefer attribute/`tfsdk` names already used by an analogous schema — see [`naming.md`](../naming.md).
- EXAMPLE: `provider/imagemirror` (`image_mirror_state.go`). Prefer a package closest to your feature over copying unrelated CRUD from that package.

## Collections and nulls

WHEN writing List, Set, or Map attributes into the model/state (including populate helpers used by Create/Read/Update):
- MUST: Set an explicit Framework value — a populated collection or a typed null (e.g. `types.ListNull(types.StringType)`) — never leave the field as Go `nil`.
- MUST: When the API omits or clears a value, assign the matching typed null (or the empty value that matches the attribute contract) so refresh does not keep stale data.
- EXAMPLE: HCP cluster `log_forwarder_ids` initialized with `types.ListNull` when none are configured (`provider/clusterrosa/hcp`).
- NOTE: Nested-object rules when Required children are missing from the API, and map clear-on-shrink, live in Create/Read/Update — see those step docs and pointers from [`resource-schema.md`](resource-schema.md).
