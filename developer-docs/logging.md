# Logging

Router: [`AGENTS.md`](../AGENTS.md). Diagnostics versus logs: [`errors.md`](errors.md). Secrets in logs: [`security.md`](security.md).

Shared by **resources and data sources**. Practitioner-facing failures use diagnostics ([`errors.md`](errors.md)); this file is provider `tflog` only.

WHEN your situation matches one of these, open **only** that section:
- WHEN adding or changing provider log calls → [Provider logging](#provider-logging)
- WHEN choosing what fields or values to include → [What to log](#what-to-log)
- DEFAULT: Open both sections for any new log call.

## Provider logging

WHEN writing provider logs (CRUD, Read, helpers):
- MUST: Use `tflog` (`Debug` / `Info` / `Warn` / `Error`) with the request `context.Context`.
- MUST NOT: Use `fmt.Print*` (or similar stdout/stderr prints) for provider debugging or operational logs.
- MUST NOT: Use diagnostics as a substitute for debug/trace detail — see [`errors.md`](errors.md).
- DEFAULT: Prefer `tflog.Debug` for routine create/read/update/delete milestones; use `Warn`/`Error` in logs only when matching an analogous package.
- DEFAULT: For new types, follow HashiCorp plugin logging (https://developer.hashicorp.com/terraform/plugin/log/managing) where it does not conflict with this repo’s rules. When editing an existing package, match that package.
- EXAMPLE: `provider/imagemirror` — `tflog.Debug` on create/update/delete with `cluster_id` and `id` fields.
- EXAMPLE: `provider/logforwarder` — Debug on CRUD milestones; Warn when not-found on Read (alongside `RemoveResource`).

## What to log

WHEN adding structured fields or message text:
- MUST: Prefer non-secret identifiers (`cluster_id`, resource `id`, operation names) that help troubleshoot.
- MUST NOT: Log secrets or sensitive values — follow [`security.md`](security.md).
- DEFAULT: Match the nearest analogous package for field names and verbosity; keep messages short.
- EXAMPLE: `provider/imagemirror` — `map[string]any{"cluster_id": ..., "id": ...}` on Debug success paths.
