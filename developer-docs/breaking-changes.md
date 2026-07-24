# Breaking changes

Router: [`AGENTS.md`](../AGENTS.md). Schema surface: [`resources/resource-schema.md`](resources/resource-schema.md). Type/attribute names: [`naming.md`](naming.md).

Catalog of user-visible contract risks and how to deprecate before a hard break. Shared by **resources and data sources**.

WHEN your situation matches one of these, open **only** that section:
- WHEN judging whether a change is breaking → [What counts as breaking](#what-counts-as-breaking)
- WHEN a breaking change is unavoidable → [When a break is necessary](#when-a-break-is-necessary)
- WHEN removing or replacing schema without an immediate hard break → [Schema deprecations](#schema-deprecations)
- WHEN unsure if human review is required → [Human-in-the-loop](#human-in-the-loop)
- DEFAULT: Prefer non-breaking changes; open [What counts as breaking](#what-counts-as-breaking) first.

## What counts as breaking

Treat any of the following as potentially breaking unless proven otherwise:

- Attribute rename, removal, or type change.
- Attribute/block/type deprecation and later removal.
- Required versus optional/computed contract changes.
- Behavioral changes in plan/apply that alter existing successful configurations.
- Import/state format changes.
- Provider-wide dependency changes that can impact runtime, generated docs, authentication, or API compatibility.

## When a break is necessary

WHEN a breaking change is necessary:
- MUST: Use the PR template **Breaking Changes** section and migration fields.
- MUST: Add tests that show the impact; update docs/examples as needed.
- MUST: Request human review before merge.
- DEFAULT: Prefer non-breaking changes; document migration when breaking is unavoidable.

## Schema deprecations

Framework guide: https://developer.hashicorp.com/terraform/plugin/framework/deprecations

WHEN removing or replacing a practitioner-facing attribute, block, or type without an immediate hard break:
- MUST: Prefer Framework deprecation (`DeprecationMessage` on the schema element) in a **minor** release so existing configs keep working with a clear warning.
- MUST: Remove the deprecated surface only in a **major** release (or other explicitly versioned break), with migration notes in the PR **Breaking Changes** section.
- MUST NOT: Silently rename or remove in a normal feature PR — see also [`naming.md`](naming.md).
- DEFAULT: If unsure whether a change needs deprecation versus an immediate break, treat it as breaking and follow [Human-in-the-loop](#human-in-the-loop).

## Human-in-the-loop

WHEN any of the following occurs, stop and request explicit human review before merge:
- Dependency bump for provider-wide tooling/runtime with broad impact (e.g. `terraform-plugin-docs`, AWS SDK modules, Terraform Plugin Framework/core).
- Schema or state model changes affecting existing resources/data sources.
- New feature appears unsupported or ambiguous in ROSA docs or OCM API.
- Security-sensitive behavior (auth, token handling, trust bundles, proxy, logging of request/response data).
- CI failures that suggest cross-repo or infrastructure issues rather than isolated code defects.

DEFAULT: Prefer non-breaking changes; document migration when breaking is unavoidable.
