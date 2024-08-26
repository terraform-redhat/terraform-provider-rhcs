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

package classic

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("Group membership creation", func() {
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on groups is check
		// that the cluster is ready, so we always need to prepare the server to respond to
		// that:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `{
				  "id": "123",
				  "name": "my-cluster",
				  "state": "ready"
				}`),
			),
		)
	})

	It("fails if group is empty", func() {
		Terraform.Source(`
		resource "rhcs_group_membership" "my_membership" {
				cluster = ""
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring(`The argument "group" is required, but no definition was found`)
	})

	It("fails if user is empty", func() {
		Terraform.Source(`
		resource "rhcs_group_membership" "my_membership" {
				group = "my-group"
				cluster = ""
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring(`The argument "user" is required, but no definition was found`)
	})

	It("fails if cluster is empty", func() {
		Terraform.Source(`
		resource "rhcs_group_membership" "my_membership" {
				group = "my-group"
				user = "my-user"
				cluster = ""
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring(`Attribute cluster cluster ID may not be empty/blank string`)
	})

	It("Can create a group membership", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/groups/dedicated-admins/users",
				),
				VerifyJSON(`{
				  "kind": "User",
				  "id": "my-admin"
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-admin"
				}`),
			),
		)

		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_group_membership" "my_membership" {
		    cluster   = "123"
		    group     = "dedicated-admins"
		    user      = "my-admin"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		// Check the state:
		resource := Terraform.Resource("rhcs_group_membership", "my_membership")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.group", "dedicated-admins"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-admin"))
		Expect(resource).To(MatchJQ(".attributes.user", "my-admin"))
	})
})
