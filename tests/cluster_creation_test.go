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

package tests

***REMOVED***
	"context"
***REMOVED***
	"os"
	"time"

	. "github.com/onsi/ginkgo"       // nolint
***REMOVED***       // nolint
	. "github.com/onsi/gomega/ghttp" // nolint

	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Cluster creation", func(***REMOVED*** {
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
	}***REMOVED***

	AfterEach(func(***REMOVED*** {
		// Stop the server:
		server.Close(***REMOVED***

		// Remove the server CA file:
		err := os.Remove(ca***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates a simple cluster", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// First thing the provider should do is search for an existing cluster with
			// the name given in the configuration:
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyFormKV("search", "name = 'my-cluster'"***REMOVED***,
				VerifyFormKV("size", "1"***REMOVED***,
				RespondWithJSON(
					http.StatusOK,
					`{
					  "kind": "ClusterList",
					  "page": 1,
					  "size": 0,
					  "total": 0,
					  "items": []
			***REMOVED***`,
				***REMOVED***,
			***REMOVED***,

			// Then the provider should send the request to create the cluster:
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
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
			***REMOVED***`,
				***REMOVED***,
				RespondWithJSON(
					http.StatusCreated,
					`{
					  "kind": "Cluster",
					  "id": "123",
					  "name": "my-cluster",
					  "state": "ready"
			***REMOVED***`,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		result := NewCommand(***REMOVED***.
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source  = "localhost/redhat/ocm"
				    }
				  }
		***REMOVED***

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}"***REMOVED***
		***REMOVED***

				resource "ocm_cluster" "my_cluster" {
				  name           = "my-cluster"
				  cloud_provider = "aws"
				  cloud_region   = "us-west-1"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", ca,
			***REMOVED***.
			Args(
				"apply",
				"-auto-approve",
			***REMOVED***.
			Run(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
