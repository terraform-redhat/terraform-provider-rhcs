# Contributing to RHCS Terraform Provider
The RHCS (Red Hat Cloud Services) provider is maintained by a team within Red Hat.
This document contains instructions about how to contribute and how the RHCS provider works.
It was targeted to developers who want to help improve the provider, and does not intended for users of the provider.

Please read this document and follow this guide to ensure your contribution will be accepted as fast as possible.

## Contributing Code Guidelines
To begin with, we appreciate your enthusiasm for contributing to RHCS Provider.
If you have any questions or uncertainties, feel free to reach out for help.

### 1. Report an issue
You can report us issues, documentation, bug fixes, code and feature requests.
Issues can be open [here](https://github.com/terraform-redhat/terraform-provider-rhcs/issues), as well as feature requests. Please look though the list of open issues before opening a new
issue. If you find an open issue you have encountered and can provide additional information, please feel free to join the conversation.

Code contributions are done by opening a pull request (PR). Please be sure that your PR is linked to an open issue/feature-request.

### 2. Development Environment
* [terraform](https://www.terraform.io/) - We are using the latest version of terraform (1.6.x)  
* [golang](https://go.dev/) - use the version declared in `go.mod` to build the provider plugin
  - You also need to correctly setup a `GOPATH`, as well as adding `$GOPATH/bin` to your `$PATH`.
  - Fork and clone the repository to `$GOPATH/src/github.com/terraform-redhat/terraform-provider-rhcs` by the `git clone` command
  - Create a new branch `git switch -c <branch-name>`
  - Run `go mod download` to download all the modules in the dependency graph.
  - Install [pre-commit](https://pre-commit.com/#install)
  - BEFORE YOUR FIRST COMMIT IN A NEW CLONE, YOU MUST RUN `make install-hooks`.
  - Try building the project by running `make build`

### 3. Make your changes with a Coding Style
Use the repository formatting helpers before committing:

```shell
make fmt         # formats Go import order and syntax plus Terraform files under examples/ and tests/, then fails if rewrites were needed
make fmt-staged  # formats staged Go import order and syntax plus staged Terraform files under examples/ and tests/, then fails if rewrites were needed
make fmt-check   # verifies Go import order/formatting plus Terraform formatting without rewriting files
make lint        # runs the pinned golangci-lint v2 configuration used by CI
make docs-lint  # runs Vale with the repo's custom style only (inclusive terminology); see .vale.ini
```

Keep the code clean and readable. Functions should be concise, exit the function as early as possible.
Best coding standards for golang can be found [here](https://go.dev/doc/effective_go).

### Required local hooks

This repository uses [pre-commit](https://pre-commit.com/) to manage git hooks. BEFORE YOUR FIRST COMMIT IN A CLONE, YOU MUST:

1. Install `pre-commit` following the [official installation guide](https://pre-commit.com/#install)
2. Run `make install-hooks` to configure the hooks

YOU MUST LET THE LOCAL HOOKS RUN ON EVERY COMMIT AND PUSH. DO NOT BYPASS LOCAL HOOKS.

The hooks are configured in `.pre-commit-config.yaml` and perform:

- `pre-commit`: formats staged Go files with `gci` + `gofmt` plus staged Terraform files under `examples/` and `tests/`, adds Apache 2.0 license headers to staged files missing them, and blocks the commit if files were rewritten so you can review and stage the updates
- `commit-msg`: validates the commit message format (JIRA-123 | type(scope): message)
- `pre-push`: runs the same steps as `make pre-push-checks` (format-check, build, generated-files check, lint, docs-lint, license-check, subsystem registry check, and `make test`)
- `pre-push` runs against committed content and blocks when staged or unstaged tracked changes are present
- check runs are fail-fast: execution stops at the first failing step

To manually run all hooks on all files:
```shell
pre-commit run --all-files
```

This covers only the `pre-commit` stage. The `commit-msg` hook cannot be exercised manually because it requires a commit message file that git creates during `git commit` — without it the script exits 0 unconditionally. The `pre-push` stage can be run explicitly:
```shell
pre-commit run --all-files --hook-stage push
```

To update hook versions:
```shell
pre-commit autoupdate
```

#### Migrating from the legacy `.githooks` system

If you previously set `core.hooksPath=.githooks` in a local clone, run `make install-hooks` — it will automatically unset that configuration and install the pre-commit hooks in its place. No manual cleanup is needed.

### 4. Test your changes and use the CI (Pre Merge)

The provider uses four automated test layers before merge:

| Layer | Command | What it exercises |
|-------|---------|-------------------|
| **Unit** | `make unit-test` | Validators, plan modifiers, mapping, and helpers in `provider/` and `internal/` |
| **Subsystem** | `make subsystem-test` | Terraform plan/apply against a mock OCM API (`subsystem/`) |
| **Utils** | `make e2e-unit-test` | Unit tests for e2e harness helpers under `tests/utils/` |
| **E2e** | `make e2e_test` | Real clusters on OpenShift CI — **not** required for every PR |

Run unit, subsystem, and utils tests locally with:

```shell
make test
```

#### Pre-merge requirements

`make pre-push-checks` (also run by the pre-push git hook and GitHub Actions) verifies:

- Formatting, build, generated files, lint, docs-lint, and license headers
- **`make check-subsystem-registry`** — every registered resource and data source type must be referenced in `subsystem/` tests, or listed in `hack/subsystem-registry-allowlist.yaml` with a ticket and reason; **new types** added on the branch must include a subsystem test
- **`make test`** — unit, subsystem, and utils suites pass

See `AGENTS.md` for when to add unit versus subsystem tests.

#### When to add which test

| Change | Required test |
|--------|----------------|
| New or changed **resource or data source** | **Subsystem** test under `subsystem/classic/` or `subsystem/hcp/` |
| New or changed **validation, plan modifiers, or helpers** (Go code) | **Unit** test in the same package (`*_test.go`) |
| **Schema / ConfigValidators** (plan-time errors) | **Unit** and/or **one** subsystem test expecting plan/apply failure — avoid duplicating the same cases in both layers |

Unit tests for validators and helpers are required by review policy and the PR testing checklist; they are **not** enforced by an automated coverage percentage gate.

#### Optional coverage tools (not merge gates)

These commands help locally; CI and pre-push do not enforce them:

| Command | Purpose |
|---------|---------|
| `make unit-test-coverage` | Package-level unit coverage for `provider/` and `internal/`; produces `coverage.out` for `go tool cover -html=coverage.out` |
| `make coverage-changed-files` | Changed-line unit coverage compared to merge base with `main` (gocovdiff, 80% threshold). **Does not include subsystem tests.** Useful before large refactors; not required to merge |

Pre-merge quality for provider behavior relies on **`make test`** (unit + subsystem + utils) and **`make check-subsystem-registry`** for registered types.

Use these commands before pushing:

```shell
make basic-checks      # convenience flow: starts with make fmt and may stop after rewrites so you can review/stage
make pre-push-checks   # exact non-mutating verification used by the pre-push hook
```

`make basic-checks` runs format, format-check, build, generated-files verification, lint, docs-lint (Vale), subsystem registry check, and unit/subsystem/utils tests.
`make lint` uses the repo's pinned `golangci-lint` v2 configuration.
`make docs-lint` runs the pinned [Vale](https://docs.vale.sh/) CLI with only the custom inclusive-language rules under `styles/InclusiveLanguage/` (general Vale styles and packages are not used). Building Vale uses `CGO_ENABLED=1` and requires a C compiler toolchain on the first install.

### 5. Manual testing and debugging using the locally compiled RHCS Provider binary
Manual testing should be performed before opening a PR to ensure there isn't any regression behavior in the provider. You can find [here an example for that](https://github.com/terraform-redhat/terraform-rhcs-rosa/tree/main/examples/rosa-classic-public-with-unmanaged-oidc)
After compiling the RHCS provider, debugging terraform provider can be difficult. But here are a some tips to make your life easier.

First, Make sure you are using your local build of the provider. `make install` will compile the project and place the binary in the local `~/.terraform/` folder.
You can then use that build in your manifests by pointing the provider to that location as such:
```
terraform {
  required_providers {
    rhcs = {
      source  = "terraform.local/local/rhcs" # points the provider to your local build
      version = ">= 1.1.0"
    }
  }
}
```
Use the `tflog` for println debugging:
```go
    tflog.Debug(ctx, msg)
```
Set environment variable `TF_LOG` to one of the log levels (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`). This will result in a more verbose output that can help you identify issues.

### 6. Commit Messages

Be sure to practice good git commit hygiene as you make your changes. All but the smallest changes should be broken up
into a few commits that tell a story. Use your git commits to provide context for the folks who will review PR. We strive
to follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary).

The commit message should follow this template:
```shell
[JIRA-TICKET] | [TYPE][(scope)][!]: <MESSAGE>

[optional BODY]

[optional FOOTER(s)]
```

Supported JIRA ticket formats: `OCM-XXXXX` or `ROSAENG-XXXX`

The commit contains the following structural types, to communicate your intent:

- `fix:` a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning).
- `feat:` a commit of the type feat introduces a new feature to the codebase (this correlates with MINOR in Semantic
  Versioning).

Types other than `fix:` and `feat:` are allowed:
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to our CI configuration files and scripts
- `docs`: Documentation only changes
- `perf`: A code change that improves performance
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- `test`: Adding missing tests or correcting existing tests

[!IMPORTANT]
DCO Sign-off Required: Every commit must include a Developer Certificate of Origin (DCO) sign-off line (Signed-off-by: Name <email>). Use git commit -s when committing.

### 7. Release Process and Changelog Automation

The changelog is automatically generated using [git-cliff](https://git-cliff.org/) configured via `cliff.toml`. Only the `CHANGELOG.md` in the `main` branch contains the complete changelog history.

**Workflow:**
1. Push a release tag (`v1.7.3`) to trigger the automation
2. GitHub Actions automatically generates the changelog from the previous release tag
3. A PR is created to `main` with the new changelog entry, labeled `changelog` to be reviewed.

The changelog follows the existing format with sections for FEATURES, ENHANCEMENTS (with Bug fixes and Documentation subsections), and other categories. Commits are automatically grouped based on their conventional commit type.

**Manual Changelog Generation:**
```bash
# Generate changelog for a specific release range
git-cliff <previous-tag>..<current-tag> --prepend CHANGELOG.md

# Example:
git-cliff v1.7.2..v1.7.3 --prepend CHANGELOG.md
```

## Related Documentation Links
* [RHCS rosa module](https://github.com/terraform-redhat/terraform-rhcs-rosa) - for creating ROSA clusters much more easily.
* [RHCS rosa HCP module](https://github.com/terraform-redhat/terraform-rhcs-rosa-hcp/) -  for creating ROSA HCP clusters much more easily.
* [ROSA project](https://docs.openshift.com/rosa/welcome/index.html) - RedHat Openshift Service on AWS (ROSA)
* [OpenShift Cluster Management API](https://api.openshift.com/) - Since the RHCS provider uses OpenShift APIs.
* [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) documentation. RHCS provider is leveraging this framework heavily.
* [Debugging Terraform](https://developer.hashicorp.com/terraform/internals/debugging) - More info about Terraform Logging and Debugging
* [Terraform Language Documentation](https://developer.hashicorp.com/terraform/language) - Information about Terraform resources and data sources.
