# Docs and examples

Published registry pages live under **`docs/`** and are often **generated**. Commands: [`CONTRIBUTING.md`](../CONTRIBUTING.md).

WHEN changing resource/data source schema or descriptions:
- MUST NOT: Hand-edit generated pages under `docs/` when the workflow requires regeneration.
- MUST: Edit sources (`templates/`, schema descriptions) and regenerate per `CONTRIBUTING.md`.
- MUST: Verify generated files are committed when regeneration updates them.

WHEN adding a new resource or data source:
- MUST: Add a runnable example under `examples/resources/<name>/` or `examples/data-sources/<name>/` (e.g. `example_1.tf`).
- MUST: Match HashiCorp Terraform style and `rhcs_*` naming; use placeholders — no real secrets.

DEFAULT: Prefer existing template and example patterns in this repo.
