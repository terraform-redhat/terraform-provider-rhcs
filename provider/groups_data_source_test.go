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

var _ = Describe("Groups data source", func(***REMOVED*** {
	It("Can list groups", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/groups"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 1,
				  "items": [
				    {
				      "id": "dedicated-admins"
				    }
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_groups" "my_groups" {
		    cluster = "123"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_groups", "my_groups"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items |length`, 1***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].id`, "dedicated-admins"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items[0].name`, "dedicated-admins"***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
