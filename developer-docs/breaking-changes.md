# Breaking changes

Treat any of the following as potentially breaking unless proven otherwise:

- Attribute rename, removal, or type change.
- Required versus optional/computed contract changes.
- Behavioral changes in plan/apply that alter existing successful configurations.
- Import/state format changes.
- Provider-wide dependency changes that can impact runtime, generated docs, authentication, or API compatibility.

WHEN a breaking change is necessary:
- MUST: Use the PR template **Breaking Changes** section and migration fields.
- MUST: Add tests that show the impact; update docs/examples as needed.
- MUST: Request human review before merge.

## Human-in-the-loop

MUST: Stop and request explicit human review before merge when any of the following occurs:

- Dependency bump for provider-wide tooling/runtime with broad impact (e.g. `terraform-plugin-docs`, AWS SDK modules, Terraform Plugin Framework/core).
- Schema or state model changes affecting existing resources/data sources.
- New feature appears unsupported or ambiguous in ROSA docs or OCM API.
- Security-sensitive behavior (auth, token handling, trust bundles, proxy, logging of request/response data).
- CI failures that suggest cross-repo or infrastructure issues rather than isolated code defects.

DEFAULT: Prefer non-breaking changes; document migration when breaking is unavoidable.
