# Agent Guide — terraform-provider-rhcs

This repository contains the Terraform provider for Red Hat Cloud Services (RHCS), with primary focus on ROSA APIs exposed through OCM.

This guide defines how coding agents should make decisions, implement changes, and decide when to escalate to a human reviewer.

## Where Rules Live

| File | Purpose |
|------|---------|
| `.cursor/rules/` | Hard, stable coding and behavior guardrails (language, testing, and workflow expectations). |
| `CONTRIBUTING.md` | Command and procedure authority (formatting, lint, verification, tests, release process). |
| `.github/pull_request_template.md` | Required PR narrative and verification checklist inputs. |
| `AGENTS.md` | Process narrative for agents: decision flow, safety triggers, and definition of done. |

Thin entrypoints `CLAUDE.md` and `GEMINI.md` should only point here to avoid drift.

If this file conflicts with `CONTRIBUTING.md` or `.cursor/rules/`, follow `CONTRIBUTING.md` first, then `.cursor/rules/`.

## HashiCorp Terraform Skills

Upstream skills live under:

- https://github.com/hashicorp/agent-skills/tree/main/terraform

Recommended skills for this provider:

- `terraform-provider-development` for provider resource/data source implementation patterns and Plugin Framework best practices.
- `terraform-test` for Terraform configuration test authoring.
- `terraform-style-guide` for Terraform HCL in examples, tests, and docs snippets.

How to use:

1. Open the relevant skill `SKILL.md`.
2. Apply the workflow rather than inventing a new one.
3. Keep this repository's `CONTRIBUTING.md` and local rules as authority.

## Product Boundaries (ROSA + OCM)

- Do not implement provider support for capabilities that are not available in ROSA.
- Validate feature availability against ROSA documentation and OCM API behavior before coding.
- When unsure whether support is ROSA HCP vs Classic (or both), stop and verify before implementation.

Reference sources:

- ROSA docs: https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/4/html/about/welcome-index
- OCM API plugin references: https://github.com/openshift-online/rosa-claude-plugins/tree/main/ocm-api

## Hard Rules

- Never hardcode secrets, API keys, tokens, kubeconfigs, AWS credentials, or customer identifiers in code, tests, docs, examples, or logs.
- Never bypass local hooks or verification gates in agent-authored guidance.
- Never introduce breaking changes silently: call out impact and migration explicitly.
- Never change generated docs manually when generation is required by repo workflow; edit source and regenerate.
- Prefer existing patterns already used in this repo over net-new architecture or naming.
- Keep error messages actionable and consistent with project standards.

## Common Implementation Flows

### Add or change a resource

1. Confirm capability exists in ROSA and OCM API.
2. Inspect analogous resource implementations in `provider/` and follow schema/state/CRUD conventions.
3. Add or update schema and state structs using existing package patterns.
4. Implement create/read/update/delete behavior and diagnostics.
5. Add/update unit tests in `provider/.../*_test.go`.
6. Add/update subsystem coverage under `subsystem/` for behavior changes.
7. Update docs source (`templates/` and schema descriptions), regenerate docs, and verify generated files.

### Add or change a data source

1. Confirm capability exists and is queryable through OCM APIs used by provider.
2. Follow existing data source patterns for schema, read behavior, and diagnostics.
3. Validate shape and computed attributes against analogous data sources.
4. Add/update unit tests and subsystem tests for the primary query flow.
5. Regenerate and validate docs for the new/changed data source.

## Breaking Change Policy

Treat any change below as potentially breaking unless proven otherwise:

- Attribute rename/removal/type change.
- Required vs optional/computed contract changes.
- Behavioral changes in plan/apply that alter existing successful configurations.
- Import/state format changes.
- Provider-wide dependency changes that can impact runtime, generated docs, authentication, or API compatibility.

When a breaking change is necessary:

1. Document it in PR under "Breaking Changes" with explicit migration steps.
2. Add tests that demonstrate old behavior impact and new expected behavior.
3. Update docs and examples to show migration path.
4. Request human review explicitly before merge.

## Human-in-the-Loop Triggers

Stop and request explicit human review before merge when any of the following occurs:

- Dependency bump for provider-wide tooling/runtime with possible broad impact, including:
  - `terraform-plugin-docs`
  - AWS SDK modules (for example `aws-sdk-go-v2`)
  - Terraform Plugin Framework/core dependencies
- Schema or state model changes affecting existing resources/data sources.
- New feature appears unsupported or ambiguous in ROSA docs or OCM API.
- Security-sensitive behavior changes (auth, token handling, trust bundles, proxy behavior, logging of request/response data).
- CI failures suggest cross-repo or infrastructure issues rather than isolated code defects.

## Use Existing Patterns First

Before introducing new structure, identify and reuse:

- Similar resource/data source package layouts.
- Existing CRUD and diagnostics patterns.
- Existing subsystem test wiring and assertions.
- Existing docs template style and generated docs workflow.

If no suitable pattern exists, document why a new pattern is needed in the PR description.

Reference implementation history:

- PR #957 can be used as a practical example of adding provider surface plus docs/tests flow: https://github.com/terraform-redhat/terraform-provider-rhcs/pull/957

## Definition of Done (Agent Checklist)

Use this checklist before handing work to humans. It is aligned with `.github/pull_request_template.md`.

- [ ] Problem statement and user impact are clearly documented.
- [ ] PR summary explains both what changed and why.
- [ ] Relevant Jira/GitHub issues and related PRs are linked.
- [ ] Change type is identified correctly (`feat`, `fix`, `docs`, etc.).
- [ ] Breaking change status is explicitly marked; migration plan provided when applicable.
- [ ] Code follows existing repository patterns and local rules.
- [ ] Required tests were added or updated for behavior changes.
- [ ] Local verification steps pass (including formatting/lint/build/pre-push checks defined in `CONTRIBUTING.md`).
- [ ] Docs were updated/regenerated when schema or behavior changed.
- [ ] Risks, limitations, and follow-up work are explicitly documented.
- [ ] Human review was requested for any trigger listed in this guide.

## Practical Review Heuristics for Agents

- Prefer small, reviewable commits scoped to one behavior change.
- Keep PR descriptions concrete with reproducible validation steps.
- Avoid speculative refactors while implementing functional changes.
- If a change touches both Classic and HCP behavior, call out parity or intentional divergence explicitly.
