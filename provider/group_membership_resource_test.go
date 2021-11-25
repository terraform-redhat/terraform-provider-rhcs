/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

***REMOVED***
	"context"
***REMOVED***
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"                         // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Group membership creation", func(***REMOVED*** {
	var ctx context.Context
	var server *Server
	var ca string
	var token string

	BeforeEach(func(***REMOVED*** {
		// Create a contet:
		ctx = context.Background(***REMOVED***

		// Create an access token:
		token = MakeTokenString("Bearer", 10*time.Minute***REMOVED***

		// Start the server:
		server, ca = MakeTCPTLSServer(***REMOVED***

		// The first thing that the provider will do for any operation on grups is check
		// that the cluster is ready, so we always need to prepare the server to respond to
		// that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "123",
				  "name": "my-cluster",
				  "state": "ready"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}***REMOVED***

	AfterEach(func(***REMOVED*** {
		// Stop the server:
		server.Close(***REMOVED***

		// Remove the server CA file:
		err := os.Remove(ca***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create a group membership", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/groups/dedicated-admins/users",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "User",
				  "id": "my-admin"
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-admin"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		result := NewTerraformRunner(***REMOVED***.
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
		***REMOVED***

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}"***REMOVED***
		***REMOVED***

				resource "ocm_group_membership" "my_membership" {
				  cluster   = "123"
				  group     = "dedicated-admins"
				  user      = "my-admin"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_group_membership", "my_membership"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.group", "dedicated-admins"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-admin"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.user", "my-admin"***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
