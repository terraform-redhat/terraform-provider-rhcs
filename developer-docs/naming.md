# Naming

Router: [`AGENTS.md`](../AGENTS.md). Package/file layout: [`packages/package-layout.md`](packages/package-layout.md). Contract changes: [`breaking-changes.md`](breaking-changes.md).

Shared by resources and data sources.

WHEN your situation matches one of these, open **only** that section:
- WHEN choosing a Terraform type name (`rhcs_*`) → [Type names](#type-names)
- WHEN naming attributes or nested blocks → [Attribute and block names](#attribute-and-block-names)
- WHEN renaming an existing type or attribute → [Breaking renames](#breaking-renames)
- DEFAULT: Open [Type names](#type-names) for a new type; [Attribute and block names](#attribute-and-block-names) for schema fields.

## Type names

WHEN choosing or changing a Terraform type name:
- MUST: Use nouns; each type is one managed object or one query result shape.
- MUST: Prefix with `rhcs_` (e.g. `rhcs_cluster_rosa_hcp`, `rhcs_versions`).
- MUST: Prefer OCM/ROSA terminology over invented names.
- MUST: Data sources — **plural** when returning a list; **singular** when returning one object (e.g. `rhcs_machine_types` versus `rhcs_info`).
- WHEN Classic and HCP need distinct types: MUST follow existing in-repo naming patterns (`*_rosa_classic`, `*_hcp`, or a shared type) — do not invent a new scheme.
- DEFAULT: Match the nearest analogous type name in `provider/`.

## Attribute and block names

WHEN naming attributes or nested blocks:
- MUST: Lowercase snake_case (e.g. `aws_account_id`, `domain_prefix`).
- MUST: Singular nouns for scalars; plural for list, set, or map.
- MUST: Boolean names so **true** means enable/do; avoid double-negative semantics.
- MUST: Nested blocks use a singular noun even when multiple instances are allowed (e.g. `sts`, `proxy`, `admin_credentials`).
- MUST: Write-only arguments use a `_wo` suffix when that matches existing provider patterns.
- DEFAULT: Prefer attribute names already used by an analogous schema in this repo over synonyms.

## Breaking renames

WHEN renaming a type or attribute:
- MUST: Follow [`breaking-changes.md`](breaking-changes.md).
- MUST NOT: Silently rename in a normal feature PR.
