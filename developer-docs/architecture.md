# Architecture

Router: [`AGENTS.md`](../AGENTS.md). Package paths: [`packages/package-layout.md`](packages/package-layout.md).

ROSA/OCM product boundaries and Classic versus HCP. Shared by resources and data sources.

WHEN your situation matches one of these, open **only** that section:
- WHEN confirming ROSA/OCM support or Classic versus HCP scope → [Scope](#scope)
- WHEN aligning with ROSA CLI behavior → [ROSA CLI parity](#rosa-cli-parity)
- DEFAULT: Open [Scope](#scope) first for any new capability or architecture-sensitive change.

## Scope

WHEN implementing or changing provider capabilities:
- MUST: Implement only capabilities available in **ROSA** (via OCM APIs used by this provider). MUST NOT add provider support for features ROSA does not offer.
- MUST: Confirm whether a change applies to **Classic**, **HCP**, or both before implementing.
- MUST: Keep Classic and HCP implementations separated. Paths and package layout: [`packages/package-layout.md`](packages/package-layout.md).
- MUST: Call out parity or intentional divergence in the PR when a change touches both Classic and HCP.
- DEFAULT: When unsure whether support is HCP, Classic, or both — stop and verify before implementation.

## ROSA CLI parity

WHEN the change corresponds to ROSA CLI behavior (including a companion `openshift/rosa` PR):
- MUST: Prefer parity with the CLI for business rules, validation, naming, and user-facing messages/logs when the provider can express the same behavior.
- MUST NOT: Copy CLI patterns that break Terraform desired-state semantics (for example CLI partial update versus full config reconciliation). Document intentional divergence in the PR.
- DEFAULT: If there is no CLI counterpart, follow ROSA docs and OCM API behavior; do not invent attributes or rules.

**See:**

- ROSA docs: https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/
- OCM API model: https://github.com/openshift-online/ocm-api-model
- ROSA CLI: https://github.com/openshift/rosa
