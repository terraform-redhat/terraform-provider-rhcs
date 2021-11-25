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

	It("Can list cloud providers types", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "kind": "CloudProviderList",
				  "page": 1,
				  "size": 7,
				  "total": 7,
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

				data "ocm_cloud_providers" "my_providers" {
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "my_providers"***REMOVED***

		// Check the GCP cloud provider:
		gcpItems, err := JQ(`.attributes.items[] | select(.name == "gcp"***REMOVED***`, resource***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		Expect(gcpItems***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		gcpItem := gcpItems[0]
		Expect(gcpItem***REMOVED***.To(MatchJQ(".id", "gcp"***REMOVED******REMOVED***
		Expect(gcpItem***REMOVED***.To(MatchJQ(".name", "gcp"***REMOVED******REMOVED***
		Expect(gcpItem***REMOVED***.To(MatchJQ(".display_name", "GCP"***REMOVED******REMOVED***

		// Check the AWS machine type:
		awsItems, err := JQ(`.attributes.items[] | select(.name == "aws"***REMOVED***`, resource***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		Expect(awsItems***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		awsItem := awsItems[0]
		Expect(awsItem***REMOVED***.To(MatchJQ(".id", "aws"***REMOVED******REMOVED***
		Expect(awsItem***REMOVED***.To(MatchJQ(".name", "aws"***REMOVED******REMOVED***
		Expect(awsItem***REMOVED***.To(MatchJQ(".display_name", "AWS"***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
