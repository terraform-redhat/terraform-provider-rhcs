# Welcome
The RHCS (Red Hat Cloud Services) provider is maintained by a team within Red Hat.
This document contains instructions about how to contribute and how the RHCS provider works.
It was targeted to developers who wish to help improve the provider, and does not intended for users of the provider.

## Contribute
To begin with, we appreciate your enthusiasm for contributing to RHCS Provider. 
If you have any questions or uncertainties, feel free to reach out for help.

You can report us issues, documentation, bug fixes, code and feature requests.
Issues can be open [here](https://github.com/terraform-redhat/terraform-provider-rhcs/issues), as well as feature requests. Please look though the list of open issues before opening a new
issue. If you find an open issue you have encountered and can provide additional information, please feel free to join the conversation.

Code contributions are done by opening a pull request (PR). Please be sure that your PR is linked to an open issue/feature-request.

> **Note:** this guide is still a work-in-progress. If you find incorrect or missing information, we'd really appreciate a PR!

Information about Terraform (resources and data sources for example) can be found in the [Terraform Language Documentation](https://developer.hashicorp.com/terraform/language)

Please read this document and follow this guid to ensure your contribution will be accepted as fast as possible.


### 1. Development Environment
* [terraform](https://www.terraform.io/) - We are using the latest version of terraform (1.6.x)  
* [golang](https://go.dev/) - version 1.18.x (to build the provider plugin)
  - You also need to correctly setup a `GOPATH`, as well as adding `$GOPATH/bin` to your `$PATH`.
  - Fork and clone the repository to `$GOPATH/src/github.com/hashicorp/terraform-provider-rhcs` by the `git clone` command
  - Create a new branch `git switch -c <brnach-name>`
  - Run `go mod download` to download all the modules in the dependency graph.
  - Try building the project by running `make build`

### 2. Make your changes with a Coding Style
Use `gofmt` in order to format your code.
Keep the code clean and readable. Functions should be concise, exit the function as early as possible.
Best coding standards for golang can be found [here](https://go.dev/doc/effective_go).

### 3. Test your changes and use the CI (Pre Merge)
We are holding three types of tests that must pass for a PR to finally be accepted and merged: 
* `unit-tests` - for testing a small unit of resource or data source functionality [here is an example of cluster_rosa_classic unit-tests](provider/clusterrosaclassic/cluster_rosa_classic_resource_test.go)
* `subsystem test` - write a test that describing what you will fix and locate it in the [subsystem test directory](subsystem).
* `end-to-end tests` - those tests simulate a real resources and run in official OpenShift CI platform.
Both `unit-tests` and `subsystem`, can be run locally before submitting a PR, by running `make tests`.

### 4. Manual testing and debugging using the locally compiled RHCS Provider binary
Manual testing should be performed before opening a PR in order to make sure there isn't any regression behavior in the provider. You can find [here an example for that](https://github.com/terraform-redhat/terraform-rhcs-rosa/tree/main/examples/rosa-classic-with-unmanaged-oidc)  
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



For more information see [Debugging Terraform](https://developer.hashicorp.com/terraform/internals/debugging)

### 5. Related Documentation Links
* [RHCS rosa module](https://github.com/terraform-redhat/terraform-rhcs-rosa) - for creating ROSA clusters much more easily.
* [ROSA project](https://docs.openshift.com/rosa/welcome/index.html) - RedHat Openshift Service on AWS (ROSA)
* [OpenShift Cluster Management API](https://api.openshift.com/) - Since the RHCS provider uses OpenShift API's.
* [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) documentation. RHCS provider is leveraging this framework heavily.

