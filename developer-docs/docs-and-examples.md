# Docs and examples

Router: [`AGENTS.md`](../AGENTS.md). Commands: [`CONTRIBUTING.md`](../CONTRIBUTING.md). Secrets in examples: [`security.md`](security.md).

Published registry pages live under **`docs/`** and are often **generated**.

WHEN your situation matches one of these, open **only** that section:
- WHEN changing schema descriptions or generated registry docs → [Generated docs](#generated-docs)
- WHEN adding a new resource or data source type → [Examples](#examples)
- DEFAULT: Open [Generated docs](#generated-docs) for schema/description changes; [Examples](#examples) for new types.

## Generated docs

WHEN changing resource/data source schema or descriptions:
- MUST NOT: Hand-edit generated pages under `docs/` when the workflow requires regeneration.
- MUST: Edit sources (`templates/`, schema descriptions) and regenerate per `CONTRIBUTING.md`.
- MUST: Verify generated files are committed when regeneration updates them.

## Examples

WHEN adding a new resource or data source:
- MUST: Add a runnable example under `examples/resources/<name>/` or `examples/data-sources/<name>/` (e.g. `example_1.tf`).
- MUST: Match HashiCorp Terraform style and `rhcs_*` naming; use placeholders — no real secrets ([`security.md`](security.md)).
- DEFAULT: Prefer existing template and example patterns in this repo.
