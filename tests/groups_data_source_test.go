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

package tests

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"                         // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Groups data source", func() {
	var ctx context.Context
	var server *Server
	var ca string
	var token string

	BeforeEach(func() {
		// Create a contet:
		ctx = context.Background()

		// Create an access token:
		token = MakeTokenString("Bearer", 10*time.Minute)

		// Start the server:
		server, ca = MakeTCPTLSServer()
	})

	AfterEach(func() {
		// Stop the server:
		server.Close()

		// Remove the server CA file:
		err := os.Remove(ca)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Can list groups", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/groups"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 1,
				  "items": [
				    {
				      "id": "dedicated-admins"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		result := NewTerraformRunner().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				data "ocm_groups" "my_groups" {
				  cluster = "123"
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_groups", "my_groups")
		Expect(resource).To(MatchJQ(`.attributes.items |length`, 1))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "dedicated-admins"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "dedicated-admins"))
	})
})
