# Resource overview

Router: [`AGENTS.md`](../../AGENTS.md). Package layout: [`package-layout.md`](../packages/package-layout.md).

Open the linked files for rules — do not treat this file as the rulebook.

DEFAULT: Prefer the nearest analogous package under `provider/` for layout and behavior shape (see [`package-layout.md`](../packages/package-layout.md)).

## WHEN adding a new resource

Follow these steps **in order**. For each step, open **only** that step’s doc (and any link it requires), implement, then move on. Do **not** preload every linked file before coding.

**Once before coding (foundations):**

1. [`architecture.md`](../architecture.md) — ROSA/OCM boundaries, Classic versus HCP, CLI parity
2. [`package-layout.md`](../packages/package-layout.md) — (+ [`package-common.md`](../packages/package-common.md) if sharing helpers)
3. [`naming.md`](../naming.md)

**Then per step (one doc at a time):**

4. [`resource-register.md`](resource-register.md)
5. [`resource-configure.md`](resource-configure.md)
6. [`resource-schema.md`](resource-schema.md)
7. [`resource-model.md`](resource-model.md)
8. [`resource-create.md`](resource-create.md)
9. [`resource-read.md`](resource-read.md)
10. [`resource-update.md`](resource-update.md)
11. [`resource-delete.md`](resource-delete.md)
12. [`resource-import.md`](resource-import.md) — if import is supported or required
13. [`errors.md`](../errors.md) — diagnostics, warn versus error, API failure and state
14. [`logging.md`](../logging.md)
15. [`testing.md`](../testing.md)
16. [`docs-and-examples.md`](../docs-and-examples.md)
17. [`security.md`](../security.md) / [`breaking-changes.md`](../breaking-changes.md) — as applicable

## WHEN changing an existing resource

Open **only** what matches the diff (plus fan-out where noted). Checking call sites for shared helpers is required by [`package-common.md`](../packages/package-common.md), not by opening every consumer doc here.

WHEN your change matches one of these, open **only** that doc first (then any fan-out named in the bullet):
- WHEN changing package layout or adding files → [`package-layout.md`](../packages/package-layout.md) (+ [`package-common.md`](../packages/package-common.md) if sharing)
- WHEN changing shared helpers → [`package-common.md`](../packages/package-common.md) (search and update call sites per that doc)
- WHEN renaming types or attributes → [`naming.md`](../naming.md) (+ [`breaking-changes.md`](../breaking-changes.md) if renaming)
- WHEN changing registration → [`resource-register.md`](resource-register.md)
- WHEN changing provider client / Configure → [`resource-configure.md`](resource-configure.md)
- WHEN changing schema / validators / plan modifiers → [`resource-schema.md`](resource-schema.md) (+ [`resource-model.md`](resource-model.md); create/read/update that use those fields; [`security.md`](../security.md) if secrets/Sensitive)
- WHEN changing model / state structs → [`resource-model.md`](resource-model.md) (+ schema if shape drifted; CRUD that maps fields)
- WHEN changing Create only → [`resource-create.md`](resource-create.md) (+ [`errors.md`](../errors.md) if changing failure handling)
- WHEN changing Read / refresh only → [`resource-read.md`](resource-read.md) (+ [`errors.md`](../errors.md) if changing failure handling)
- WHEN changing Update only → [`resource-update.md`](resource-update.md) (+ [`errors.md`](../errors.md) if changing failure handling)
- WHEN changing Delete only → [`resource-delete.md`](resource-delete.md) (+ [`errors.md`](../errors.md) if changing failure handling)
- WHEN changing Import only → [`resource-import.md`](resource-import.md) (+ [`errors.md`](../errors.md) if changing failure handling)
- WHEN changing CRUD only (no schema change) → matching create/read/update/delete (+ [`errors.md`](../errors.md) if changing failure handling)
- WHEN changing diagnostics / API error handling → [`errors.md`](../errors.md)
- WHEN changing logs → [`logging.md`](../logging.md)
- WHEN changing tests → [`testing.md`](../testing.md)
- WHEN changing docs / examples → [`docs-and-examples.md`](../docs-and-examples.md)
- DEFAULT: Do not open sibling lifecycle docs unless the change forces it.

WHEN the change may alter the user-visible contract or plan/apply behavior for existing configurations:
- MUST: Read [`breaking-changes.md`](../breaking-changes.md) and follow the PR **Breaking Changes** section if applicable.
- MUST NOT: Duplicate the breaking-change catalog in resource step docs — that file is the single list.
