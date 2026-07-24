# Agent guide — terraform-provider-rhcs

Source of truth for AI assistants and review tooling (including CodeRabbit). Rules live in **`developer-docs/`**; commands in **`CONTRIBUTING.md`**.

Terraform provider for Red Hat Cloud Services (RHCS), focused on ROSA APIs via OCM.

## Where to look

| Topic | When to read | Doc |
|-------|--------------|-----|
| Architecture | Any change; Classic versus HCP; ROSA/OCM boundaries | [`developer-docs/architecture.md`](developer-docs/architecture.md) |
| Resources and data sources | Adding or changing provider types | [`developer-docs/resources-and-datasources.md`](developer-docs/resources-and-datasources.md) |
| Testing | Unit, subsystem, e2e, registry | [`developer-docs/testing.md`](developer-docs/testing.md) |
| Security | Secrets, Sensitive, Trivy | [`developer-docs/security.md`](developer-docs/security.md) |
| Breaking changes | Schema/behavior/deps; human review | [`developer-docs/breaking-changes.md`](developer-docs/breaking-changes.md) |
| Docs and examples | Generated docs, templates, examples | [`developer-docs/docs-and-examples.md`](developer-docs/docs-and-examples.md) |
| Commands and PR checks | Before opening a PR | [`CONTRIBUTING.md`](CONTRIBUTING.md) |
| PR submission checklist | Before opening a PR | [`.github/pull_request_template.md`](.github/pull_request_template.md) |

Entrypoints [`CLAUDE.md`](CLAUDE.md) and [`GEMINI.md`](GEMINI.md) point here.

**Precedence:** `CONTRIBUTING.md` (commands) > `developer-docs/` (rules) > this file (routing). HashiCorp skills lose to `CONTRIBUTING.md` / `developer-docs/` on conflict.

## Skills

[HashiCorp terraform skills](https://github.com/hashicorp/agent-skills/tree/main/terraform) — **`CONTRIBUTING.md` / `developer-docs/` win** on conflict.

| Skill | When |
|-------|------|
| **terraform-provider-development** | Plugin Framework CRUD/schema mechanics |
| **terraform-test** | Acceptance / test patterns |
| **terraform-style-guide** | HCL in `examples/` and test manifests |

## Workflow

1. Confirm ROSA/OCM support — [`developer-docs/architecture.md`](developer-docs/architecture.md).
2. Resource or data source? — [`developer-docs/resources-and-datasources.md`](developer-docs/resources-and-datasources.md); follow an analogous package under `provider/`.
3. Tests — [`developer-docs/testing.md`](developer-docs/testing.md); commands in [`CONTRIBUTING.md`](CONTRIBUTING.md).
4. Docs/examples — [`developer-docs/docs-and-examples.md`](developer-docs/docs-and-examples.md).
5. Security — [`developer-docs/security.md`](developer-docs/security.md).
6. Breaking or high-risk? — [`developer-docs/breaking-changes.md`](developer-docs/breaking-changes.md).
7. Before PR — [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`.github/pull_request_template.md`](.github/pull_request_template.md).

## Guardrails

- MUST NOT: Document bypassing local hooks or verification gates.
- MUST: Prefer existing patterns; keep commits small and reviewable.
- MUST: Keep PR descriptions concrete with reproducible validation steps.
- MUST NOT: Speculate refactors while implementing functional changes.

## CI container images

Dockerfiles and Go bumps: [CI container images and Go version bumps](CONTRIBUTING.md#ci-container-images-and-go-version-bumps) in **`CONTRIBUTING.md`**.

| Dockerfile | Built by | Used for |
|------------|----------|----------|
| `Dockerfile` | Konflux | Shipping provider binary |
| `Dockerfile.clients` | Prow | Presubmit: `make pre-push-checks` |
| `build/ci-tf-e2e.Dockerfile` | Prow | E2E runner (`rhcs-tf-e2e`) |
