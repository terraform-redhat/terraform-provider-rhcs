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

var _ = Describe("Cloud providers data source", func(***REMOVED*** {
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

	It("Can list cloud providers", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    },
				    {
				      "id": "gcp",
				      "name": "gcp",
				      "display_name": "GCP"
				    }
				  ]
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

				data "ocm_cloud_providers" "all" {
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "all"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 2***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].id`, "aws"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].name`, "aws"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].display_name`, "AWS"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].id`, "gcp"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].name`, "gcp"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].display_name`, "GCP"***REMOVED******REMOVED***
	}***REMOVED***

	It("Can search cloud providers", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"***REMOVED***,
				VerifyFormKV("search", "display_name like 'A%'"***REMOVED***,
				VerifyFormKV("order", "display_name asc"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    },
				    {
				      "id": "azure",
				      "name": "azure",
				      "display_name": "Azure"
				    }
				  ]
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

				data "ocm_cloud_providers" "a" {
				  search = "display_name like 'A%'"
				  order  = "display_name asc"
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "a"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.search`, "display_name like 'A%'"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.order`, "display_name asc"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 2***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].id`, "aws"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].name`, "aws"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].display_name`, "AWS"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].id`, "azure"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].name`, "azure"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].display_name`, "Azure"***REMOVED******REMOVED***
	}***REMOVED***

	It("Populates `item` if there is exactly one result", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 1,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    }
				  ]
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

				data "ocm_cloud_providers" "a" {
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "a"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item.id`, "aws"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item.name`, "aws"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item.display_name`, "AWS"***REMOVED******REMOVED***
	}***REMOVED***

	It("Doesn't populate `item` if there are zero results", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 0,
				  "total": 0,
				  "items": []
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

				data "ocm_cloud_providers" "all" {
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "all"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item`, nil***REMOVED******REMOVED***
	}***REMOVED***

	It("Doesn't populate `item` if there are multiple results", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    },
				    {
				      "id": "gcp",
				      "name": "gcp",
				      "display_name": "GCP"
				    }
				  ]
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

				data "ocm_cloud_providers" "all" {
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "all"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item`, nil***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
