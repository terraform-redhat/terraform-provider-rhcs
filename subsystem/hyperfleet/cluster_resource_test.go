/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hyperfleet

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_cluster_hyperfleet", func() {
	Context("Provider not configured for hyperfleet", func() {
		It("Reports a descriptive error when hyperfleet_url is absent from the provider", func() {
			// The suite configures the provider with OCM credentials only
			// (no hyperfleet_url / aws_account_id). The resource Configure()
			// detects the missing hyperfleet client and surfaces a clear
			// diagnostic telling the user what to add to their provider block.
			Terraform.Source(`
				resource "rhcs_cluster_hyperfleet" "test" {
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
					release_image         = "quay.io/openshift-release-dev/ocp-release:4.16.0-multi"
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("hyperfleet_url")
		})
	})

	Context("Provider configured with hyperfleet_url but missing aws_account_id", func() {
		It("Reports a descriptive error when aws_account_id is absent", func() {
			// hyperfleet_url is set but aws_account_id is omitted. The provider
			// Configure() detects the missing account ID and surfaces a clear
			// diagnostic before any API call is attempted.
			Terraform.Source(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
					release_image         = "quay.io/openshift-release-dev/ocp-release:4.16.0-multi"
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("aws_account_id")
		})
	})

	Context("Provider configured with mismatched aws_region", func() {
		It("Reports a descriptive error when aws_region contradicts the region in hyperfleet_url", func() {
			// An aliased provider block carries hyperfleet_url anchored in us-east-1
			// but aws_region explicitly set to us-west-2. The provider Configure()
			// detects the contradiction and errors before any API call is made.
			// OCM credentials are intentionally absent; Configure() tolerates that
			// when hyperfleet_url is present (hyperfleet-only mode).
			Terraform.Source(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
					aws_account_id = "123456789012"
					aws_region     = "us-west-2"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
					release_image         = "quay.io/openshift-release-dev/ocp-release:4.16.0-multi"
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("AWS region mismatch")
		})
	})
})
