# Package layout

Router: [`AGENTS.md`](../../AGENTS.md). Type names: [`naming.md`](../naming.md). Classic/HCP product rules: [`architecture.md`](../architecture.md). Shared helpers: [`package-common.md`](package-common.md). Examples: [`docs-and-examples.md`](../docs-and-examples.md).

WHERE packages and files live under `provider/` (and related test trees).

WHEN your situation matches one of these, open **only** that section:
- WHEN adding or reshaping a package’s files → [Package and files](#package-and-files)
- WHEN a type differs by Classic versus HCP → [Classic versus HCP](#classic-versus-hcp)
- WHEN placing package-local helpers → [Package-local helpers](#package-local-helpers)
- DEFAULT: Open [Package and files](#package-and-files) first; follow [`package-common.md`](package-common.md) for shared helpers.

## Package and files

- MUST: Keep the resource (or data source) and its state/model in the **same** package directory.
- MUST: One primary Terraform type per package (resource, data source, or resource + related data source in that package).
WHEN adding a **new** package:
- MUST: Use `*_resource.go` for Schema/CRUD/Configure, `*_state.go` for `tfsdk` model structs, and `*_datasource.go` (or `*_data_source.go` if matching a close neighbor) for data sources.
WHEN changing an **existing** package:
- MUST: Follow that package’s existing file naming (`resource.go` / `state.go` / `datasource.go` is acceptable if already used there).
- MUST NOT: Rename files for style-only churn in a behavior PR.
- DEFAULT: Copy layout from the nearest analogous package under `provider/`.

## Classic versus HCP

WHEN a type differs by architecture:
- MUST: Place implementations under `provider/<feature>/classic/` and/or `provider/<feature>/hcp/` as applicable.
- MUST: Mirror coverage under `subsystem/classic/` and/or `subsystem/hcp/` (see [`testing.md`](../testing.md)).
- DEFAULT: If unsure which tree — stop; see [`architecture.md`](../architecture.md).

## Package-local helpers

WHEN validators or helpers apply to **only** this package:
- MUST: Keep them in the same package (e.g. `*_validators.go`, `helpers.go`).
WHEN they are shared across packages or Classic/HCP of a feature:
- MUST: Follow [`package-common.md`](package-common.md) — search for an existing helper before adding a new one.

NOTE: Runnable examples for new types — follow [`docs-and-examples.md`](../docs-and-examples.md).
