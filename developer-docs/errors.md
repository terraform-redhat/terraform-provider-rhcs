# Errors

Router: [`AGENTS.md`](../AGENTS.md). Offline config validation (not API failures): [`resources/resource-schema.md`](resources/resource-schema.md) (resources) / data sources: [`datasources/datasource-overview.md`](datasources/datasource-overview.md). Day 2 product design (separate resource versus cluster embed): [`resources/resource-create.md`](resources/resource-create.md). Logging (`tflog`): [`logging.md`](logging.md). Secrets in logs and diagnostics: [`security.md`](security.md).

Terraform surfaces **diagnostics** (errors and warnings). This file covers API/apply/refresh outcomes and how they interact with state — shared by **resources and data sources**. It is **not** schema validators or ConfigValidators.

Data sources: use [Severity](#severity), [Messages](#messages), and the **Read**-oriented rules under [Lifecycle and state](#lifecycle-and-state). Create / Update / Delete / Import rows apply to resources.

WHEN your situation matches one of these, open **only** that section:
- WHEN choosing error versus warning for an API or apply failure → [Severity](#severity)
- WHEN handling not-found, forbidden, timeouts, or how errors affect state → [Lifecycle and state](#lifecycle-and-state)
- WHEN writing diagnostic text or failure logs → [Messages](#messages)
- DEFAULT: Open [Lifecycle and state](#lifecycle-and-state) for RPC-specific behavior; use [Severity](#severity) when unsure whether to warn or error.

## Severity

Framework diagnostics: https://developer.hashicorp.com/terraform/plugin/framework/diagnostics

WHEN returning diagnostics after an API or apply/refresh failure:
- MUST: Use an **error** when the provider could not complete the intended operation and practitioners must not assume success.
- MUST: Use a **warning** only for informative or suboptimal cases that must not fail the operation (Framework: warnings do not halt execution).
- MUST NOT: Use diagnostics as a substitute for `tflog` debug detail — see [`logging.md`](logging.md).
- DEFAULT: Prefer **error** for OCM/API failures unless matching an **existing** package behavior you are editing.
- WHEN editing an existing cluster attribute that already embeds Day 2 behavior and warns on failure: DEFAULT: Match that package’s warn-versus-error — do not rewrite for style. Product placement for Day 2 (separate resource versus cluster embed): [`resources/resource-create.md`](resources/resource-create.md).
- WHEN failure would leave operators believing a safety control is active when it is not (e.g. delete protection): MUST use an error — do not warn-and-continue.

## Lifecycle and state

DEFAULT: Match the nearest analogous package when behavior is package-specific.

WHEN implementing **Create** and the API fails:
- MUST: Return an **error** for typical failures.
- MUST: When create would duplicate an existing remote object, error so users import instead of creating twice (when the API allows detecting that).
- NOTE: Multi-step Create — errors do not block persisting **returned** state; prefer saving the API-returned partial state so the resource can be managed or destroyed rather than abandoned (see HashiCorp diagnostics and analogous cluster packages).

WHEN implementing **Read** (resource or data source) and the remote object is gone (e.g. OCM HTTP 404):
- MUST: For **resources**, call `RemoveResource` and return early — **no** error diagnostic.
- MUST: For **data sources**, return an **error** (or match the analogous data source) — data sources are not removed from state the same way.
- EXAMPLE (resource Read not-found): `provider/imagemirror` — 404 → `RemoveResource`.

WHEN implementing **Read** and refresh/lookup fails for any other reason (timeout, 5xx, authz, parse, network):
- MUST: Return an **error** — for resources, Terraform **keeps prior state**.
- MUST NOT: Call `RemoveResource` on timeout, 5xx, or ambiguous failures.

WHEN implementing **Update** and the API fails (including not-found):
- MUST: Return an **error** — Terraform **keeps prior state**.
- MUST NOT: Silently remove the resource from state on Update not-found (unless an analogous package documents a deliberate exception).

WHEN implementing **Delete** and the remote object is already gone (e.g. 404):
- MUST: Treat as **success** — do not return an error.
- EXAMPLE: `provider/imagemirror` Delete — 404 returns early with no error diagnostic (Framework drops state).

WHEN implementing **Delete** and delete fails for any other reason:
- MUST: Return an **error** — Terraform **keeps the resource in state**.
- NOTE: Established **state-only remove** exceptions (warn + drop state when the API cannot delete while the parent exists) live in [`resources/resource-delete.md`](resources/resource-delete.md#state-only-remove) — do not invent new soft-delete paths for ordinary API failures.

NOTE: MUST NOT call `State.Set` on a successful Delete — Framework removes state automatically when no error diagnostics are returned. Calling `RemoveResource` explicitly is also acceptable (but not required on the success path).

WHEN implementing **Import** and lookup fails:
- MUST: Return an **error** (import fails) — see [`resources/resource-import.md`](resources/resource-import.md).

WHEN status is **403** / unauthorized (or similar authz failures):
- MUST: Return an **error** — do not treat as “resource gone.”

WHEN the call **times out** or the wait helper fails:
- MUST: Return an **error** — do not treat as not-found / `RemoveResource`.
- MUST: When the timeout comes from a configurable schema attribute (e.g. `destroy_timeout`, `max_cluster_wait_timeout_in_minutes`, `max_hcp_cluster_wait_timeout_in_minutes`), point practitioners at that attribute: prefer `AddAttributeError` on that path when practical; otherwise name the attribute in the detail (reuse package constants / helpers when they exist).
- DEFAULT: Match the nearest analogous package (Classic/HCP siblings) and ROSA CLI wording.
- EXAMPLE: `provider/clusterrosa` destroy / wait timeout messaging (`provider/clusterrosa/common/consts.go` and cluster Create/Delete wait paths).

## Messages

WHEN creating error or warning diagnostics:
- MUST: Prefer message text and severity consistent with **ROSA CLI** (and with the nearest in-repo Classic/HCP sibling) when the provider can express the same behavior.
- MUST: Search for an existing shared constant or helper (package or `provider/.../common`) before inventing a new summary/detail string — reuse when it fits.
- MUST: Use a short practitioner-oriented **summary** and a **detail** with enough context to troubleshoot (operation, ids when safe, API status or message, underlying error text when safe) — not only a generic “API failed” summary.
- MUST NOT: Put secrets or sensitive values in diagnostic summaries/details or `tflog` fields — follow [`security.md`](security.md).
- MUST: Prefer ids and non-secret identifiers in messages; redact or omit sensitive request/response bodies.
- MUST: Append framework/`Get`/`Set` diagnostics to the response; check `HasError()` before continuing.
- MUST: When logging on the failure path with `tflog`, include the same class of troubleshooting context without secrets — see [`logging.md`](logging.md).
- MUST NOT: Put offline validation rules here — see resource or data source schema docs.
- DEFAULT: Prefer existing patterns in the analogous package over inventing new wording or severity. Follow HashiCorp diagnostics guidance for summary/detail quality (https://developer.hashicorp.com/terraform/plugin/framework/diagnostics).
