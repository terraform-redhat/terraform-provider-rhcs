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
	"strings"
	"time"

	. "github.com/onsi/ginkgo"                         // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Cluster creation", func(***REMOVED*** {
	var ctx context.Context
	var server *Server
	var ca string
	var token string

	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	const template = `{
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
	    "compute": 3,
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
	  },
	  "state": "ready"
	}`

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

	It("Creates basic cluster", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusCreated, template***REMOVED***,
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

				resource "ocm_cluster" "my_cluster" {
				  name           = "my-cluster"
				  cloud_provider = "aws"
				  cloud_region   = "us-west-1"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Saves API and console URLs to the state", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithJSON(http.StatusCreated, template***REMOVED***,
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

				resource "ocm_cluster" "my_cluster" {
				  name           = "my-cluster"
				  cloud_provider = "aws"
				  cloud_region   = "us-west-1"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cluster", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.api_url", "https://my-api.example.com"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.console_url", "https://my-console.example.com"***REMOVED******REMOVED***
	}***REMOVED***

	It("Saves console URL to the state", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithJSON(http.StatusCreated, template***REMOVED***,
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

				resource "ocm_cluster" "my_cluster" {
				  name           = "my-cluster"
				  cloud_provider = "aws"
				  cloud_region   = "us-west-1"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cluster", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.api_url", "https://my-api.example.com"***REMOVED******REMOVED***
	}***REMOVED***

	It("Sets compute nodes and machine type", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
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
				    "compute": 3,
				    "compute_machine_type": {
				      "kind": "MachineType",
				      "id": "r5.xlarge"
				    }
				  }
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusCreated, template***REMOVED***,
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

				resource "ocm_cluster" "my_cluster" {
				  name                 = "my-cluster"
				  cloud_provider       = "aws"
				  cloud_region         = "us-west-1"
				  compute_nodes        = 3
				  compute_machine_type = "r5.xlarge"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cluster", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.compute_nodes", 3.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.compute_machine_type", "r5.xlarge"***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails if the cluster already exists", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithJSON(http.StatusBadRequest, `{
				  "id": "400",
				  "code": "CLUSTERS-MGMT-400",
				  "reason": "Cluster 'my-cluster' already exists"
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

				resource "ocm_cluster" "my_cluster" {
				  name           = "my-cluster"
				  cloud_provider = "aws"
				  cloud_region   = "us-west-1"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
