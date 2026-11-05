// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

// The operator roles resource and data source talk directly to AWS IAM (like
// `rosa create operator-roles` in seaplane mode), so these subsystem tests
// exercise schema and plan wiring only — no AWS credentials are available in
// the stubbed subsystem environment. Full CRUD acceptance coverage runs in
// the e2e suites against real accounts.
var _ = Describe("rhcs_seaplane_operator_roles", func() {
	const issuerURL = "https://oidc.s3.us-east-1.amazonaws.com/subsystem-test"

	It("plans the full operator role set for a prefix", func() {
		// The resource does not call the Platform API, but the provider must
		// still configure successfully, so point seaplane_url at the mock
		// server.
		seaplaneProvider := seaplaneProviderConfig()
		Terraform.Source(seaplaneProvider + `
			resource "rhcs_seaplane_operator_roles" "roles" {
				provider              = rhcs.sp
				operator_roles_prefix = "subsystem-op-roles"
				oidc_issuer_url       = "` + issuerURL + `"
			}
		`)

		By("validating the configuration")
		Expect(Terraform.Validate().ExitCode).To(Equal(0))

		By("planning the role set")
		plan := Terraform.Run("plan", "-no-color")
		Expect(plan.ExitCode).To(Equal(0))
		plan.VerifyOutputContainsSubstring("1 to add")
	})

	It("rejects a non-https OIDC issuer at plan time", func() {
		seaplaneProvider := seaplaneProviderConfig()
		Terraform.Source(seaplaneProvider + `
			resource "rhcs_seaplane_operator_roles" "roles" {
				provider              = rhcs.sp
				operator_roles_prefix = "subsystem-op-roles"
				oidc_issuer_url       = "http://oidc.example.com/test"
			}
		`)

		plan := Terraform.Run("plan", "-no-color")
		Expect(plan.ExitCode).ToNot(Equal(0))
	})

	It("references the operator roles data source for discovery", func() {
		// The data source reads from AWS IAM, which is not stubbed here; this
		// case keeps the data source in the subsystem registry and validates
		// its schema wiring.
		Terraform.Source(`
			data "rhcs_seaplane_operator_roles" "roles" {
				operator_roles_prefix = "subsystem-op-roles"
			}
		`)

		Expect(Terraform.Validate().ExitCode).To(Equal(0))
	})
})

// seaplaneProviderConfig returns a provider block pointing seaplane_url at
// the mock server so the provider configures successfully. The operator
// roles resource itself talks only to AWS.
func seaplaneProviderConfig() string {
	return `
		provider "rhcs" {
			alias          = "sp"
			seaplane_url   = "` + TestServer.URL() + `"
			aws_account_id = "123456789012"
			aws_region     = "us-east-1"
		}
	`
}
