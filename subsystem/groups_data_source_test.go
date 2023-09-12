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

var _ = Describe("Groups data source", func() {
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
				    },
					{
				      "id": "dedicated-admins2"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_groups" "my_groups" {
		    cluster = "123"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_groups", "my_groups")
		Expect(resource).To(MatchJQ(`.attributes.items |length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "dedicated-admins"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "dedicated-admins"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "dedicated-admins2"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].name`, "dedicated-admins2"))
	})
})
