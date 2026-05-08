# Agent Guide — terraform-provider-rhcs

This repository contains the Terraform provider for Red Hat Cloud Services (RHCS), with primary focus on ROSA APIs exposed through OCM.

This guide defines how coding agents should make decisions, implement changes, and decide when to escalate to a human reviewer.

## Where rules live (avoid drift)

Do **not** duplicate checklists or command lists here. Use one source per concern:

| Source | Use it for |
|--------|------------|
| `README.md` | Project overview, prerequisites, contributor setup summary, limitations, and links to deeper docs — **read this** for context before changing behavior or docs. |
| `docs/` | Published provider documentation (often **generated** from `templates/` and schema); do not edit generated pages by hand when the workflow requires regeneration — follow `CONTRIBUTING.md`. |
| `examples/` | Runnable Terraform under `examples/` (resources, data sources, guides paths); use for manual checks and as references for HCL style — align with provider schema and HashiCorp Terraform style. |
| `CONTRIBUTING.md` | Commands, hooks, tests, formatting, release — **procedural authority**. |
| `.github/pull_request_template.md` | What to fill in on a PR and the **Developer Verification Checklist** — **submission checklist**. |
| `.cursor/rules/` | Code style and Terraform Plugin Framework guardrails. |
| **`AGENTS.md` (this file)** | ROSA/OCM boundaries, escalation triggers, and **high-level** implementation flow — not step-by-step commands. |

Thin entrypoints `CLAUDE.md` and `GEMINI.md` should only point here to avoid drift.

**Precedence:** If anything here conflicts with `CONTRIBUTING.md` or `.cursor/rules/`, follow `CONTRIBUTING.md` first, then `.cursor/rules/`.

**Before opening a PR:** Complete the checklist in `.github/pull_request_template.md`.

## HashiCorp Terraform Skills

Upstream skills: https://github.com/hashicorp/agent-skills/tree/main/terraform

Recommended for this provider: `terraform-provider-development`, `terraform-test`, `terraform-style-guide` (open each skill’s `SKILL.md`). Treat `CONTRIBUTING.md` and `.cursor/rules/` as overriding generic skill text when they disagree.

## Product boundaries (ROSA + OCM)

- Do not implement provider support for capabilities that are not available in ROSA.
- Validate feature availability against ROSA documentation and the official OCM API behavior before implementing resource or data source code.
- When unsure whether support is ROSA HCP vs Classic (or both), stop and verify before implementation.

Reference sources:

- ROSA docs: https://docs.redhat.com/en/documentation/red_hat_openshift_service_on_aws/
- OCM API model (API specification): https://github.com/openshift-online/ocm-api-model

## Expectations for agent-produced changes

These are **review and design expectations**. They are not all enforced by automation; contributors and reviewers still apply them. Concrete commands and gates are in `CONTRIBUTING.md` (for example hooks, `make pre-push-checks`, coverage rules).

- Do not hardcode secrets, API keys, tokens, kubeconfigs, AWS credentials, or customer identifiers in code, tests, docs, examples, or logs.
- Do not document bypassing local hooks or verification gates.
- Do not ship silent breaking changes: call out impact and migration in the PR (see **Breaking Changes** in the PR template).
- Do not edit generated provider docs by hand when the workflow requires regeneration; edit sources and regenerate per `CONTRIBUTING.md`.
- Prefer existing patterns in this repo over new architecture or naming.
- Keep error messages actionable and consistent with project standards.

## Common implementation flows

### Add or change a resource

1. Confirm capability exists in ROSA and OCM API.
2. Inspect analogous resource implementations in `provider/` and follow schema/state/CRUD conventions.
3. Add or update schema and state structs using existing package patterns.
4. Implement create/read/update/delete behavior and diagnostics.
5. Add/update unit tests in `provider/.../*_test.go`.
6. Add/update subsystem coverage under `subsystem/` for behavior changes.
7. Update docs source (`templates/` and schema descriptions), regenerate docs, and verify generated files.

### Add or change a data source

1. Confirm capability exists and is queryable through OCM APIs used by the provider.
2. Follow existing data source patterns for schema, read behavior, and diagnostics.
3. Validate shape and computed attributes against analogous data sources.
4. Add/update unit tests and subsystem tests for the primary query flow.
5. Regenerate and validate docs for the new/changed data source.

## Breaking change policy

Treat any change below as potentially breaking unless proven otherwise:

- Attribute rename/removal/type change.
- Required vs optional/computed contract changes.
- Behavioral changes in plan/apply that alter existing successful configurations.
- Import/state format changes.
- Provider-wide dependency changes that can impact runtime, generated docs, authentication, or API compatibility.

When a breaking change is necessary, use the PR template’s **Breaking Changes** section and migration fields, add tests that show the impact, update docs/examples, and request human review as required by `CONTRIBUTING.md`.

## Human-in-the-loop triggers

Stop and request explicit human review before merge when any of the following occurs:

- Dependency bump for provider-wide tooling/runtime with possible broad impact, including:
  - `terraform-plugin-docs`
  - AWS SDK modules (for example `aws-sdk-go-v2`)
  - Terraform Plugin Framework/core dependencies
- Schema or state model changes affecting existing resources/data sources.
- New feature appears unsupported or ambiguous in ROSA docs or OCM API.
- Security-sensitive behavior changes (auth, token handling, trust bundles, proxy behavior, logging of request/response data).
- CI failures suggest cross-repo or infrastructure issues rather than isolated code defects.

## Trivy (IaC misconfiguration)

Repo config: root **`trivy.yaml`** (severity, scanners, skips; Terraform under **`examples/`**, **`tests/`**, **`generate_example_usages/`**, and root **`Dockerfile`**). CodeRabbit may run Trivy when enabled in **`.coderabbit.yaml`**. References: [Trivy config file](https://trivy.dev/latest/docs/references/configuration/config-file/), [filtering / ignores](https://trivy.dev/latest/docs/configuration/filtering/).

When **`trivy config`** reports a **misconfiguration** (check IDs like **`AWS-0104`**, **`DS-0002`** — not CVE vulnerability rows from **`trivy fs`** vuln scans):

1. **Prefer fixing** the HCL/Dockerfile (least privilege, encryption, IMDSv2, non-root user, etc.).
2. If an ignore is required, add **`#trivy:ignore:<id>`** on the line **immediately above** the Terraform resource or Dockerfile instruction, with a **short `#` comment** on the same line or the line above explaining why (narrow scope).
3. Use **`.trivyignore`** only when inline suppression is not possible — one ID per line with a **`#` justification** above each.

## Use existing patterns first

Before introducing new structure, identify and reuse:

- Similar resource/data source package layouts.
- Existing CRUD and diagnostics patterns.
- Existing subsystem test wiring and assertions.
- Existing docs template style and generated docs workflow.

If no suitable pattern exists, document why in the PR description.

To see how a similar change was done, search **merged** pull requests in this repository for the relevant `provider/` or `subsystem/` paths instead of relying on a single historical PR link.

## Practical review heuristics for agents

- Prefer small, reviewable commits scoped to one behavior change.
- Keep PR descriptions concrete with reproducible validation steps (align with the template’s **How to Test**).
- Avoid speculative refactors while implementing functional changes.
- If a change touches both Classic and HCP behavior, call out parity or intentional divergence explicitly.
