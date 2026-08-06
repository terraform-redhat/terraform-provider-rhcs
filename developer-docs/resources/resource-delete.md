# Resource delete

Router: [`resource-overview.md`](resource-overview.md). Diagnostics and API failures: [`errors.md`](../errors.md). Logging: [`logging.md`](../logging.md).

WHEN your situation matches one of these, open **only** that section:
- WHEN implementing Delete (standard flow: read state → call API → drop state) → [Delete method](#delete-method)
- WHEN destroy must wait until the remote object is gone (or uses a destroy timeout attribute) → [Wait for destroy](#wait-for-destroy)
- WHEN the API cannot delete the object but Terraform must still drop state → [State-only remove](#state-only-remove)
- DEFAULT: WHEN adding or changing delete behavior, open [Delete method](#delete-method) first, then any matching WHEN above.

## Delete method

WHEN implementing Delete:
- MUST: Call the delete API with values that match OCM/ROSA semantics per [`architecture.md`](../architecture.md) when the provider can express the same behavior.
- MUST: Handle not-found and all other delete failures per [`errors.md`](../errors.md).
- DEFAULT: For new types, follow HashiCorp Delete (https://developer.hashicorp.com/terraform/plugin/framework/resources/delete) where it does not conflict with this repo’s rules. When editing an existing package, match that package.
- EXAMPLE: `provider/imagemirror` Delete (state get → DELETE → 404 early return with no error; other errors → diagnostic).

NOTE: Configurable destroy duration — see [Wait for destroy](#wait-for-destroy) and the timeouts NOTE in [`resource-create.md`](resource-create.md).

## Wait for destroy

WHEN Delete must poll until the remote object is gone, or exposes a destroy wait/timeout attribute:
- MUST: Reuse existing wait helpers and timeout attributes like the nearest analogous package (e.g. cluster `destroy_timeout`).
- MUST: On wait/timeout failure, return an **error** and point practitioners at the timeout attribute when one exists — see [`errors.md`](../errors.md).
- DEFAULT: Prefer ROSA CLI / Classic–HCP sibling destroy-wait behavior when the provider can express it.
- EXAMPLE: `provider/clusterrosa` classic/HCP Delete wait paths and `destroy_timeout` messaging (`provider/clusterrosa/common/consts.go`).

## State-only remove

WHEN the remote API cannot (or must not) delete this object while the parent still exists, but practitioners need Terraform to drop state:
- MUST: Document the behavior in generated docs / guides when user-visible.
- MUST: Prefer a **warning** (state removed) only when matching an established package pattern for that case — do not invent new soft-delete paths for ordinary API failures.
- DEFAULT: Match the nearest analogous package (Classic/HCP siblings) and ROSA CLI constraints.
- EXAMPLE: `provider/defaultingress` (classic/HCP) — cannot delete default ingress; warn and `RemoveResource`.
- EXAMPLE: `provider/machinepool/classic` — last pool cannot be deleted; warn and `RemoveResource`.
