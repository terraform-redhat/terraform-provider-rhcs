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

	It("Succeeds if the cluster doesn't exist", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJSON(`{
				  "kind": "Cluster",
				  "name": "my-cluster",
				  "cloud_provider": {
				    "kind": "CloudProvider",
				    "id": "aws"
				  },
				  "region": {
				    "kind": "CloudRegion",
				    "id": "us-west-1"
				  },
				  "nodes": {
				    "compute": 10,
				    "compute_machine_type": {
				      "kind": "MachineType",
				      "id": "r5.xlarge"
				    }
				  }
				}`),
				RespondWithJSON(http.StatusCreated, `{
				  "id": "123",
				  "name": "my-cluster",
				  "cloud_provider": {
				    "id": "aws"
				  },
				  "region": {
				    "id": "us-west-1"
				  },
				  "multi_az": false,
				  "properties": {},
				  "api": {
				    "url": "https://my-api.example.com"
				  },
				  "console": {
				    "url": "https://my-console.example.com"
				  },
				  "nodes": {
				    "compute": 10,
				    "compute_machine_type": {
				      "id": "r5.xlarge"
				    }
				  },
				  "state": "ready"
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
				      source = "localhost/redhat/ocm"
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
				  nodes          = {
					  compute = 10
					  compute_machine_type = "r5.xlarge"
				  }
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_cluster", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.api_url", "https://my-api.example.com"))
		Expect(resource).To(MatchJQ(".attributes.console_url", "https://my-console.example.com"))
		Expect(resource).To(MatchJQ(".attributes.nodes.compute", 10.0))
		Expect(resource).To(MatchJQ(".attributes.nodes.compute_machine_type", "r5.xlarge"))
	})

	It("Fails if the cluster already exists", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJSON(`{
				  "kind": "Cluster",
				  "name": "my-cluster",
				  "cloud_provider": {
				    "kind": "CloudProvider",
				    "id": "aws"
				  },
				  "region": {
				    "kind": "CloudRegion",
				    "id": "us-west-1"
				  }
				}`),
				RespondWithJSON(http.StatusBadRequest, `{
				  "id": "400",
				  "code": "CLUSTERS-MGMT-400",
				  "reason": "Cluster 'my-cluster' already exists"
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
				      source = "localhost/redhat/ocm"
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
			Apply(ctx)
		Expect(result.ExitCode()).ToNot(BeZero())
	})
})
