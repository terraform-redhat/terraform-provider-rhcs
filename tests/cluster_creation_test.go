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

	. "github.com/onsi/ginkgo"       // nolint
	. "github.com/onsi/gomega"       // nolint
	. "github.com/onsi/gomega/ghttp" // nolint

	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Cluster creation", func() {
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

	It("Creates a simple cluster", func() {
		// Prepare the server:
		server.AppendHandlers(
			// First thing the provider should do is search for an existing cluster with
			// the name given in the configuration:
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"),
				VerifyFormKV("search", "name = 'my-cluster'"),
				VerifyFormKV("size", "1"),
				RespondWithJSON(
					http.StatusOK,
					`{
					  "kind": "ClusterList",
					  "page": 1,
					  "size": 0,
					  "total": 0,
					  "items": []
					}`,
				),
			),

			// Then the provider should send the request to create the cluster:
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJSON(
					`{
					  "kind": "Cluster",
					  "cloud_provider": {
					    "kind": "CloudProvider",
					    "id": "aws"
					  },
					  "name": "my-cluster",
					  "region": {
					    "kind": "CloudRegion",
					    "id": "us-west-1"
					  }
					}`,
				),
				RespondWithJSON(
					http.StatusCreated,
					`{
					  "kind": "Cluster",
					  "id": "123",
					  "name": "my-cluster",
					  "state": "ready"
					}`,
				),
			),
		)

		// Run the apply command:
		result := NewCommand().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source  = "localhost/redhat/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				resource "ocm_cluster" "my_cluster" {
				  name           = "my-cluster"
				  cloud_provider = "aws"
				  cloud_region   = "us-west-1"
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Args(
				"apply",
				"-auto-approve",
			).
			Run(ctx)
		Expect(result.ExitCode()).To(BeZero())
	})
})
