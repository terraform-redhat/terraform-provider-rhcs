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
***REMOVED***

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Versions data source", func(***REMOVED*** {
	It("Can list versions", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				VerifyFormKV("search", "enabled = 't'"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "openshift-v4.8.1",
				      "raw_id": "4.8.1"
				    },
				    {
				      "id": "openshift-v4.8.2",
				      "raw_id": "4.8.2"
				    }
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_versions" "my_versions" {
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_versions", "my_versions"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 2***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].id`, "openshift-v4.8.1"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].name`, "4.8.1"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].id`, "openshift-v4.8.2"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].name`, "4.8.2"***REMOVED******REMOVED***
	}***REMOVED***

	It("Can search versions", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				VerifyFormKV("search", "enabled = 't' and channel_group = 'fast'"***REMOVED***,
				VerifyFormKV("order", "raw_id desc"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "openshift-v4.8.1-fast",
				      "raw_id": "4.8.1"
				    },
				    {
				      "id": "openshift-v4.8.2-fast",
				      "raw_id": "4.8.2"
				    }
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_versions" "my_versions" {
		    search = "enabled = 't' and channel_group = 'fast'"
		    order  = "raw_id desc"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_versions", "my_versions"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 2***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].id`, "openshift-v4.8.1-fast"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].name`, "4.8.1"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].id`, "openshift-v4.8.2-fast"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[1].name`, "4.8.2"***REMOVED******REMOVED***
	}***REMOVED***

	It("Populates `item` if there is exactly one result", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 1,
				  "items": [
				    {
				      "id": "openshift-v4.8.1",
				      "raw_id": "4.8.1"
				    }
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_versions" "my_versions" {
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_versions", "my_versions"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item.id`, "openshift-v4.8.1"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item.name`, "4.8.1"***REMOVED******REMOVED***
	}***REMOVED***

	It("Doesn't populate `item` if there are zero results", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 0,
				  "total": 0,
				  "items": []
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_versions" "my_versions" {
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_versions", "my_versions"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item`, nil***REMOVED******REMOVED***
	}***REMOVED***

	It("Doesn't populate `item` if there are multiple results", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "openshift-v4.8.1",
				      "raw_id": "4.8.1"
				    },
				    {
				      "id": "openshift-v4.8.2",
				      "raw_id": "4.8.2"
				    }
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_versions" "my_versions" {
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_versions", "my_versions"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.item`, nil***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
