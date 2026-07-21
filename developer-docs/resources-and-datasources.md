# Resources and data sources

Use the Terraform Plugin Framework only (not the deprecated SDK). Framework CRUD/schema mechanics: HashiCorp skill **`terraform-provider-development`**. Repo rules below win on conflict.

## File layout

- MUST: Resource implementation in `*_resource.go` (Schema, Create, Read, Update, Delete; Configure when the resource needs provider-level data or clients).
- MUST: State/model structs with `tfsdk` tags in `*_state.go`.
- MUST: Data sources in `*_datasource.go` (may share `*_state.go` or define their own state types).
- MUST: Keep resource and state in the same package; one resource (or resource + data source) per package.
- MUST: Prefer an analogous existing package under `provider/` as the pattern reference.

## Naming

- MUST: Prefix type names with `rhcs_` (e.g. `rhcs_cluster_rosa_hcp`, `rhcs_versions`).
- MUST: Match OCM/ROSA terminology for type and attribute names.
- MUST: Plural data source type names when returning a list; singular when returning one object.
- MUST: Attribute and block names — lowercase with underscores; singular for scalars, plural for list/set/map; singular noun for nested blocks.
- MUST: Boolean attributes — `true` means do/enable; avoid double-negative semantics.
- MUST: Write-only arguments use a `_wo` suffix when matching existing provider patterns.

## RHCS conventions

- MUST: Log with `tflog`; MUST NOT use `fmt.Print*` for provider debugging.
- MUST: Resources that modify clusters wait with `clusterWait.WaitForClusterToBeReady` (or equivalent) where applicable.
- MUST: Use `github.com/openshift-online/ocm-sdk-go` types correctly (e.g. `cmv1`).
- MUST: Resources that support `terraform import` implement `ResourceWithImportState` and validate ID format.
- MUST: Mark secret attributes with schema **Sensitive**.
- MUST: Keep error diagnostics actionable and consistent with existing helpers.

WHEN populating nested blocks from the API (Read or after Create/Update):
- MUST: Set a nested block only when every **Required** attribute can be set from the API; otherwise leave the block unset.
DEFAULT: Follow the nearest similar resource/data source package when unsure.
