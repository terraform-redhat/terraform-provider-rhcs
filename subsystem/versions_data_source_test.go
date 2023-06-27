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

package provider

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Versions data source", func() {
	It("Can list versions", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				VerifyFormKV("search", "enabled = 't'"),
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
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_versions" "my_versions" {
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_versions", "my_versions")
		Expect(resource).To(MatchJQ(`.attributes.items | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "openshift-v4.8.1"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "4.8.1"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "openshift-v4.8.2"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].name`, "4.8.2"))
	})

	It("Can search versions", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				VerifyFormKV("search", "enabled = 't' and channel_group = 'fast'"),
				VerifyFormKV("order", "raw_id desc"),
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
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_versions" "my_versions" {
		    search = "enabled = 't' and channel_group = 'fast'"
		    order  = "raw_id desc"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_versions", "my_versions")
		Expect(resource).To(MatchJQ(`.attributes.items | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "openshift-v4.8.1-fast"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "4.8.1"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "openshift-v4.8.2-fast"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].name`, "4.8.2"))
	})

	It("Populates `item` if there is exactly one result", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
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
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_versions" "my_versions" {
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_versions", "my_versions")
		Expect(resource).To(MatchJQ(`.attributes.item.id`, "openshift-v4.8.1"))
		Expect(resource).To(MatchJQ(`.attributes.item.name`, "4.8.1"))
	})

	It("Doesn't populate `item` if there are zero results", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 0,
				  "total": 0,
				  "items": []
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_versions" "my_versions" {
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_versions", "my_versions")
		Expect(resource).To(MatchJQ(`.attributes.item`, nil))
	})

	It("Doesn't populate `item` if there are multiple results", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
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
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_versions" "my_versions" {
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_versions", "my_versions")
		Expect(resource).To(MatchJQ(`.attributes.item`, nil))
	})
})
