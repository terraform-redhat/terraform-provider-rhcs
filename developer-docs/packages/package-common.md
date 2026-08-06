# Package common helpers

Router: [`AGENTS.md`](../../AGENTS.md). Package/file placement for a single type: [`package-layout.md`](package-layout.md).

Reuse and placement of shared validators, plan modifiers, waiters, and conversion helpers.

WHEN your situation matches one of these, open **only** that section:
- WHEN adding a helper that might already exist → [Reuse before invent](#reuse-before-invent)
- WHEN deciding where shared code should live → [Where shared code goes](#where-shared-code-goes)
- WHEN changing an existing shared helper → [Existing `common` directories](#existing-common-directories)
- DEFAULT: Open [Reuse before invent](#reuse-before-invent) first.

## Reuse before invent

WHEN adding a validator, plan modifier, waiter, conversion helper, or similar:
- MUST: Search existing code first — `provider/common/` (including `attrvalidators/`, `planmodifiers/`), then `provider/<feature>/common/`, then the local package.
- MUST: Reuse or extend an existing helper when it fits.
- MUST NOT: Add a parallel helper that duplicates behavior already in those locations.
- DEFAULT: Prefer HashiCorp `terraform-plugin-framework-validators` (and existing `attrvalidators`) over a new custom validator when they cover the rule.

## Where shared code goes

WHEN helpers are used by **more than one** package under the same feature (e.g. Classic and HCP):
- MUST: Place them in `provider/<feature>/common/` (follow existing feature layout).
WHEN helpers are used across **unrelated** features:
- MUST: Place them in `provider/common/` (or a subpackage such as `attrvalidators/`, `planmodifiers/`).
WHEN helpers apply to **only one** package:
- MUST: Keep them in that package — see [`package-layout.md`](package-layout.md).
WHEN sharing a nested schema/state shape across parents:
- MUST: Prefer a dedicated package under `provider/` modeled on an analogous existing package — MUST NOT dump feature-specific schema into `provider/common/` without a clear cross-feature need.
- DEFAULT: Follow the nearest analogous shared package; do not create a new top-level `util`/`helpers` package for one-off code.

## Existing `common` directories

- MUST: Continue using established `provider/common` and `provider/<feature>/common` packages.
- MUST NOT: Rename or relocate them for style-only reasons in a behavior PR.

WHEN changing an existing shared helper, validator, or plan modifier:
- MUST: Search for call sites under `provider/` and update them in the same change.
- MUST NOT: Change shared behavior without updating callers (or documenting an intentional break).
