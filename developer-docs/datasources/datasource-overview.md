# Data source overview

Router: [`AGENTS.md`](../../AGENTS.md). Package layout: [`package-layout.md`](../packages/package-layout.md).

Open the linked files for rules — do not treat this file as the rulebook.

DEFAULT: Prefer the nearest analogous **data source** package under `provider/` for layout and behavior shape (see [`package-layout.md`](../packages/package-layout.md)).

NOTE: Data-source-specific step docs (register / schema / model / read) are **deferred** to a follow-up PR. Until then, use the shared foundations below, [`errors.md`](../errors.md) / [`logging.md`](../logging.md) for Read failures and logs, and [`resource-configure.md`](../resources/resource-configure.md) when Configure patterns apply. Do not invent parallel DS rule files in this PR.

## WHEN adding a new data source

Follow these steps **in order**. For each step, open **only** that doc (and any link it requires), implement, then move on. Do **not** preload every linked file before coding.

**Foundations (shared with resources):**

1. [`architecture.md`](../architecture.md) — ROSA/OCM boundaries, Classic versus HCP, CLI parity
2. [`package-layout.md`](../packages/package-layout.md) — (+ [`package-common.md`](../packages/package-common.md) if sharing helpers)
3. [`naming.md`](../naming.md)

**Then (until DS step docs land):**

4. Match the nearest analogous data source package under `provider/` for register / schema / model / Read / Configure shape
5. Configure (if needed) — [`resource-configure.md`](../resources/resource-configure.md) (same ProviderData patterns)
6. [`errors.md`](../errors.md) — diagnostics; data-source Read not-found is an **error** (not `RemoveResource`)
7. [`logging.md`](../logging.md)
8. [`testing.md`](../testing.md)
9. [`docs-and-examples.md`](../docs-and-examples.md)
10. [`security.md`](../security.md) / [`breaking-changes.md`](../breaking-changes.md) — as applicable

## WHEN changing an existing data source

Open **only** what matches the diff (plus fan-out where noted). Checking call sites for shared helpers is required by [`package-common.md`](../packages/package-common.md), not by opening every consumer doc here.

WHEN your change matches one of these, open **only** that doc first (then any fan-out named in the bullet):
- WHEN changing package layout or adding files → [`package-layout.md`](../packages/package-layout.md) (+ [`package-common.md`](../packages/package-common.md) if sharing)
- WHEN changing shared helpers → [`package-common.md`](../packages/package-common.md) (search and update call sites per that doc)
- WHEN renaming types or attributes → [`naming.md`](../naming.md) (+ [`breaking-changes.md`](../breaking-changes.md) if renaming)
- WHEN changing register / schema / model / Read shape → nearest analogous data source package under `provider/` (+ [`resource-configure.md`](../resources/resource-configure.md) if Configure; [`security.md`](../security.md) if secrets/Sensitive)
- WHEN changing diagnostics / API error handling → [`errors.md`](../errors.md)
- WHEN changing logs → [`logging.md`](../logging.md)
- WHEN changing tests → [`testing.md`](../testing.md)
- WHEN changing docs / examples → [`docs-and-examples.md`](../docs-and-examples.md)
- DEFAULT: Do not open unrelated docs unless the change forces it.

WHEN the change may alter the user-visible contract or read/plan behavior for existing configurations:
- MUST: Read [`breaking-changes.md`](../breaking-changes.md) and follow the PR **Breaking Changes** section if applicable.
- MUST NOT: Duplicate the breaking-change catalog in data source docs — that file is the single list.
