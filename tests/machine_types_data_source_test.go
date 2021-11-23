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

var _ = Describe("Machine types data source", func() {
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

	It("Can list machine types", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/machine_types"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 2,
				  "items": [
				    {
				      "name": "custom-16-131072-ext - Memory Optimized",
				      "category": "memory_optimized",
				      "size": "large",
				      "id": "custom-16-131072-ext",
				      "memory": {
				        "value": 137438953472,
				        "unit": "B"
				      },
				      "cpu": {
				        "value": 16,
				        "unit": "vCPU"
				      },
				      "cloud_provider": {
				        "id": "gcp"
				      },
				      "ccs_only": false,
				      "generic_name": "highmem-16"
				    },
				    {
				      "name": "c5.12xlarge - Compute optimized",
				      "category": "compute_optimized",
				      "size": "xxlarge",
				      "id": "c5.12xlarge",
				      "memory": {
				        "value": 103079215104,
				        "unit": "B"
				      },
				      "cpu": {
				        "value": 48,
				        "unit": "vCPU"
				      },
				      "cloud_provider": {
				        "id": "aws"
				      },
				      "ccs_only": true,
				      "generic_name": "highcpu-48"
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

				data "ocm_machine_types" "my_machines" {
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_machine_types", "my_machines")

		// Check the GCP machine type:
		gcpTypes, err := JQ(`.attributes.items[] | select(.cloud_provider == "gcp")`, resource)
		Expect(err).ToNot(HaveOccurred())
		Expect(gcpTypes).To(HaveLen(1))
		gcpType := gcpTypes[0]
		Expect(gcpType).To(MatchJQ(".cloud_provider", "gcp"))
		Expect(gcpType).To(MatchJQ(".id", "custom-16-131072-ext"))
		Expect(gcpType).To(MatchJQ(".name", "custom-16-131072-ext - Memory Optimized"))
		Expect(gcpType).To(MatchJQ(".cpu", 16.0))
		Expect(gcpType).To(MatchJQ(".ram", 137438953472.0))

		// Check the AWS machine type:
		awsTypes, err := JQ(`.attributes.items[] | select(.cloud_provider == "aws")`, resource)
		Expect(err).ToNot(HaveOccurred())
		Expect(awsTypes).To(HaveLen(1))
		awsType := awsTypes[0]
		Expect(awsType).To(MatchJQ(".cloud_provider", "aws"))
		Expect(awsType).To(MatchJQ(".id", "c5.12xlarge"))
		Expect(awsType).To(MatchJQ(".name", "c5.12xlarge - Compute optimized"))
		Expect(awsType).To(MatchJQ(".cpu", 48.0))
		Expect(awsType).To(MatchJQ(".ram", 103079215104.0))
	})
})
