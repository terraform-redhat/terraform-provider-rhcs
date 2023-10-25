# Welcome

The RHCS (Red Hat Cloud Services***REMOVED*** provider is a project that is developed in the open by a small team at Red Hat, and is open to contributions from the community.
This document is targeted to developers who wish to help improve the provider, and does not intended for users of the provider.

## Contribute
You (yes, you***REMOVED***, are welcome to contribute to this project in any depth you see fit. Reporting issues, documentation, bug fixes, code and feature requests.
Issues can be open [here](https://github.com/terraform-redhat/terraform-provider-rhcs/issues***REMOVED***, as well as feature requests. Please look though the list of open issues before opening a new
issue. If you find an open issue you have encountered and can provide additional information, please feel free to join the conversation.

Code contntributions are done by opening a pull request. Please be sure that your PR is linked to an open issue/feature-request.
Please read this document and follow this guid to ensure your contribution will be accepted as fast as possible.


### 1. Development Environment
you will need to install [terraform](https://www.terraform.io/***REMOVED*** and [golang](https://go.dev/***REMOVED***.
 - We are using the latest version of terraform (1.6.x***REMOVED***
 - Please ensure Go version matches the version targeted in `go.mod` at the root of the repository
 - Clone the repository and run `go mod download`

Try building the project by running `make build`

#### 1.1. Coding Style
Use `gofmt` in order to format your code. This is very important to us.
Keep the code clean and readable. Functions should be concise, exit the function as early as possible.
A good place to read about best coding standards can be found [here](https://go.dev/doc/effective_go***REMOVED***.

For more information, dont hesitate to reach out to our core engineering team for help.

### 2. Debugging
Debugging a terraform provider can be difficult. But here are a some tips to make your life easier.
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
Set environment variable `TF_LOG` to one of the log levels (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`***REMOVED***. This will result in a more verbose output that can help you identify issues.

For more information see [Debugging Terraform](https://developer.hashicorp.com/terraform/internals/debugging***REMOVED***

### 3. CI (Pre Merge***REMOVED***
Each change to the code base will go though a searies of tests. Some of these tests you can run locally before submitting a PR.
These tests are `unit-tests` and `subsystem`, You can run them both locally by running `make tests`.

Other tests are run on the official OpenShift CI platform. These tests are more rounded tests. They all must pass for a PR to finnaly be accepted and merged.

### 4. Related Documentation
Since the RHCS provider uses OpenShift API's. [OpenShift Cluster Management API](https://api.openshift.com/***REMOVED*** is a valuable resource. Espesially when writing `subsytem` tests
for your patch.

Another valuable resource is the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework***REMOVED*** documention. RHCS provider
is leveraging this framwork heavily.
