# Contributing to RHCS Terraform Provider
The RHCS (Red Hat Cloud Services) provider is maintained by a team within Red Hat.
This document contains instructions about how to contribute and how the RHCS provider works.
It was targeted to developers who wish to help improve the provider, and does not intended for users of the provider.

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
* [golang](https://go.dev/) - version 1.21.x (to build the provider plugin)
  - You also need to correctly setup a `GOPATH`, as well as adding `$GOPATH/bin` to your `$PATH`.
  - Fork and clone the repository to `$GOPATH/src/github.com/hashicorp/terraform-provider-rhcs` by the `git clone` command
  - Create a new branch `git switch -c <branch-name>`
  - Run `go mod download` to download all the modules in the dependency graph.
  - Try building the project by running `make build`

### 3. Make your changes with a Coding Style
Use `gofmt` in order to format your code. You can invoke the formatting before committing with `make fmt`.

Keep the code clean and readable. Functions should be concise, exit the function as early as possible.
Best coding standards for golang can be found [here](https://go.dev/doc/effective_go).

### 4. Test your changes and use the CI (Pre Merge)
We are holding three types of tests that must pass for a PR to finally be accepted and merged: 
* `unit-tests` - for testing a small unit of resource or data source functionality [here is an example of cluster_rosa_classic unit-tests](provider/clusterrosa/classic/cluster_rosa_classic_resource_test.go)
* `subsystem test` - write a test that describing what you will fix and locate it in the [subsystem test directory](subsystem).
* `end-to-end tests` - those tests simulate a real resources and run in official OpenShift CI platform.
Both `unit-tests` and `subsystem`, can be run locally before submitting a PR, by running `make tests`.

### 5. Manual testing and debugging using the locally compiled RHCS Provider binary
Manual testing should be performed before opening a PR in order to make sure there isn't any regression behavior in the provider. You can find [here an example for that](https://github.com/terraform-redhat/terraform-rhcs-rosa/tree/main/examples/rosa-classic-public-with-unmanaged-oidc) 
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
[JIRA-TICKET] | [TYPE]: <MESSAGE>

[optional BODY]

[optional FOOTER(s)]
```

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

### 7. Changelog

We maintain a user-friendly CHANGELOG so users can see at a glance whether a release affects them and gauge the risk of an upgrade.

We follow the Terraform Plugin [changelog specifications](https://developer.hashicorp.com/terraform/plugin/best-practices/versioning#changelog-specification).

#### Changie automation tool

This project uses [Changie](https://changie.dev/) for changelog automation.

To add a new entry to the CHANGELOG:

1. Install Changie using the [installation instructions](https://changie.dev/guide/installation/).
2. Run `changie new` and choose a **kind** of change (see categories below).
3. Fill out the body following the entry format. Changie will then prompt for a GitHub issue or pull request number.
4. Commit the generated file(s) under `.changes/unreleased/` along with your code changes.

#### Pull request types to CHANGELOG

The CHANGELOG is for operator- and user-impacting changes. Not every change needs an entry.

##### Changes that should not have a CHANGELOG entry

- Documentation updates
- Testing updates
- Code refactoring

##### Changes that may have a CHANGELOG entry

- Dependency updates: if the update includes relevant bug fixes or enhancements that affect users, mention them.

##### Changes that should have a CHANGELOG entry

The following kinds are available in Changie (see `.changie.yaml`): **FEATURES**, **ENHANCEMENTS**, **BUG FIXES**, **BREAKING CHANGES**. Choose the kind that best matches your change.

###### Features (new resources and data sources)

Use the **FEATURES** kind for new resources and new data sources. For a new resource or data source, the entry is typically the name of the resource or data source, or a short one-line description.

```markdown
rhcs_log_forwarder
```

```markdown
data source: rhcs_log_forwarders
```

For other new capabilities (e.g. new arguments or behavior on existing resources), use **ENHANCEMENTS** instead.

###### Enhancements

Use the **ENHANCEMENTS** kind for improvements and new options. Use a prefix for the resource or data source when relevant, then a colon and a brief summary. Use a `provider` prefix for provider-level changes.

```markdown
resource/rhcs_cluster_rosa_hcp: Add audit_log_arn parameter
```

###### Bug fixes

Use the **BUG FIXES** kind. Same prefix convention as enhancements (resource/data source/provider).

```markdown
resource/rhcs_cluster_rosa_classic: Fix autoscaler requiring two applies
```

###### Breaking changes and removals

Use the **BREAKING CHANGES** kind for removals or behavior changes that require user action. Include a short summary and migration guidance if applicable.

```markdown
resource/rhcs_example: Removed `deprecated_field`; use `replacement_field` instead
```

## Related Documentation Links
* [RHCS rosa module](https://github.com/terraform-redhat/terraform-rhcs-rosa) - for creating ROSA clusters much more easily.
* [RHCS rosa HCP module](https://github.com/terraform-redhat/terraform-rhcs-rosa-hcp/) -  for creating ROSA HCP clusters much more easily.
* [ROSA project](https://docs.openshift.com/rosa/welcome/index.html) - RedHat Openshift Service on AWS (ROSA)
* [OpenShift Cluster Management API](https://api.openshift.com/) - Since the RHCS provider uses OpenShift APIs.
* [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) documentation. RHCS provider is leveraging this framework heavily.
* [Debugging Terraform](https://developer.hashicorp.com/terraform/internals/debugging) - More info about Terraform Logging and Debugging
* [Terraform Language Documentation](https://developer.hashicorp.com/terraform/language) - Information about Terraform (resources and data sources for example) can be found in the
