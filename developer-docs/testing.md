# Testing

Commands live in [`CONTRIBUTING.md`](../CONTRIBUTING.md). Rules below define **when** to add which layer.

## Layers

| Layer | Where | Purpose |
|-------|--------|---------|
| Unit | `provider/.../*_test.go`, `internal/.../*_test.go` | Validators, state mapping, diagnostics text |
| Subsystem | `subsystem/classic/`, `subsystem/hcp/` | Plan/apply and negative cases against stubbed OCM |
| E2E | `tests/e2e/` | Full Terraform apply against CI profiles |
| Harness | `tests/utils/exec/`, `tests/utils/profilehandler/`, `tests/tf-manifests/`, `tests/ci/profiles/` | Args/profiles → Terraform variables and provider attributes |

## When to add tests

WHEN adding or changing a **resource or data source**:
- MUST: Add or update a subsystem test under `subsystem/classic/` or `subsystem/hcp/` for the main flow.
- MUST: Use existing patterns (`TestServer`, `Terraform` runner, Ginkgo/Gomega).
- MUST: Add or update unit tests in the same package for schema, validation, state mapping, or CRUD helpers as applicable.

WHEN changing validation, plan modifiers, helpers, error messages, or schema:
- MUST: Update matching unit and/or subsystem assertions; do not assume stale substring checks still pass.
- MUST: Prefer one primary test layer per rule (unit versus subsystem negative); see [`CONTRIBUTING.md`](../CONTRIBUTING.md).

WHEN changing how values reach Terraform (flat versus nested, tfvars keys, `ClusterArgs` HCL tags):
- MUST: Keep Classic and HCP manifests under `tests/tf-manifests/` aligned unless divergence is intentional.

WHEN changing profile fields or skip conditions (e.g. `IsAdminEnabled()`):
- MUST: Confirm newly enabled e2e cases have correct fixtures and expectations.

WHEN the change touches `tests/utils/exec` or tf-manifests:
- MUST NOT: Treat subsystem coverage alone as sufficient — review e2e harness wiring.

## Subsystem registry

- MUST: Every registered `rhcs_*` type is referenced in at least one subsystem test (`resource "rhcs_…"` or `data "rhcs_…"`).
- MUST: Run `make check-subsystem-registry` (also part of `make pre-push-checks`).
- MUST NOT: Add an allowlist entry for **new** types — add a subsystem test instead.
- MUST: Use `hack/subsystem-registry-allowlist.yaml` only for temporary, documented exceptions (`type`, `ticket`, `reason`); remove the entry in the same PR that adds coverage.
- MUST: New types on the branch (versus merge base with `main`) without a subsystem reference fail the check even if allowlisted elsewhere.

DEFAULT: Docs-only or refactor-only changes need no new subsystem tests unless behavior or registry surface changes.
