<!--
Please provide enough context so reviewers can understand:
1) the problem,
2) why this change is needed,
3) what changed,
4) how you validated it.

Use N/A when the option is not applicable to your case.

Commit format requirement:
[JIRA-TICKET] | [TYPE][(scope)][!]: <MESSAGE>
TYPE must be one of:
feat, fix, docs, style, refactor, test, chore, build, ci, perf
For details, see: ./CONTRIBUTING.md
-->

## PR Summary
<!-- brief text of 1 or 2 lines with the most important changes and outcomes -->

## Detailed Description of the Issue
<!-- Describe the root problem, scope, impact, and user/business context -->

## Related Issues and PRs
<!-- Link all tracking items and related code changes -->
- Jira: [OCM-XXXXX](https://jira.url/OCM-XXXXX)
- Fixes: `#`
- Related PR(s):
- Related design/docs:

## Type of Change
<!-- Check the primary type this PR represents -->
- [ ] feat - adds a new user-facing capability.
- [ ] fix - resolves an incorrect behavior or bug.
- [ ] docs - updates documentation only.
- [ ] style - formatting or naming changes with no logic impact.
- [ ] refactor - code restructuring with no behavior change.
- [ ] test - adds or updates tests only.
- [ ] chore - maintenance work (tooling, housekeeping, non-product code).
- [ ] build - changes build system, packaging, or dependencies for build output.
- [ ] ci - changes CI pipelines, jobs, or automation workflows.
- [ ] perf - improves performance without changing intended behavior.

## Previous Behavior
<!-- What users or systems experienced before this change -->

## Behavior After This Change
<!-- What changes now, including user-visible and non-user-visible behavior -->

## How to Test (Step-by-Step)
<!-- Provide reproducible validation instructions -->
### Preconditions
<!-- Required setup, environment, credentials, flags, cluster state, etc. -->

### Test Steps
1.
2.
3.

### Expected Results
<!-- What should happen after running the steps above -->

## Proof of the Fix
<!-- Attach evidence that demonstrates the changed behavior -->
- Screenshots:
- Videos:
- Logs/CLI output:
- Other artifacts:

## Breaking Changes
- [ ] No breaking changes
- [ ] Yes, this PR introduces a breaking change (describe impact and migration plan below; see [breaking-change guidance](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/breaking-changes.md))

### Breaking Change Details / Migration Plan
<!-- Required only when breaking changes are introduced -->

## Developer Verification Checklist
- [ ] Commit subject/title follows `[JIRA-TICKET] | [TYPE][(scope)][!]: <MESSAGE>`.
- [ ] PR description clearly explains both **what** changed and **why**.
- [ ] Relevant Jira/GitHub issues and related PRs are linked.
- [ ] `make install-hooks` has been run in this clone.
- [ ] `make pre-push-checks` passes.
- [ ] Documentation was added/updated where appropriate (see [docs and examples](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/docs-and-examples.md)).
- [ ] Any risk, limitation, or follow-up work is documented.
- [ ] **Classic and HCP:** If this PR touches both Classic and HCP paths, I called out parity or intentional divergence (see [architecture](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/architecture.md)).
- [ ] **Auth / secrets / sensitive:** Changes involving credentials, tokens, Sensitive attributes, or request/response logging follow [security](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/security.md).

### Testing (check all that apply; use N/A when not relevant)
- [ ] **N/A** — no provider resource/data source or `provider/` / `internal/` logic changes.
- [ ] **New or changed resource / data source** — subsystem test added or updated under `subsystem/classic/` or `subsystem/hcp/` (see [testing](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/testing.md) and [resources and data sources](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/resources-and-datasources.md)).
- [ ] **New or changed validation, plan modifiers, or helpers** — unit tests in the same package (`*_test.go`), or a subsystem negative test when integration-only (not both for the same cases unless a wiring smoke test is needed).
- [ ] **Schema / config validation** — unit test and/or subsystem test expecting plan/apply failure (one primary layer per rule; see [CONTRIBUTING.md](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/CONTRIBUTING.md) and [testing](https://github.com/terraform-redhat/terraform-provider-rhcs/blob/main/developer-docs/testing.md)).
- [ ] **Validators / helpers** — unit tests in the same package where applicable (review policy; no automated coverage % gate).
- [ ] `make check-subsystem-registry` passes.
- [ ] I manually tested the change when behavior is user-visible.
