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

var _ = Describe("Machine types data source", func(***REMOVED*** {
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

	It("Can list machine types", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/machine_types"***REMOVED***,
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

				data "ocm_machine_types" "my_machines" {
		***REMOVED***
				`,
				"URL", server.URL(***REMOVED***,
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"***REMOVED***,
			***REMOVED***.
			Apply(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := result.Resource("ocm_machine_types", "my_machines"***REMOVED***

		// Check the GCP machine type:
		gcpTypes, err := JQ(`.attributes.items[] | select(.cloud_provider == "gcp"***REMOVED***`, resource***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		Expect(gcpTypes***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		gcpType := gcpTypes[0]
		Expect(gcpType***REMOVED***.To(MatchJQ(".cloud_provider", "gcp"***REMOVED******REMOVED***
		Expect(gcpType***REMOVED***.To(MatchJQ(".id", "custom-16-131072-ext"***REMOVED******REMOVED***
		Expect(gcpType***REMOVED***.To(MatchJQ(".name", "custom-16-131072-ext - Memory Optimized"***REMOVED******REMOVED***
		Expect(gcpType***REMOVED***.To(MatchJQ(".cpu", 16.0***REMOVED******REMOVED***
		Expect(gcpType***REMOVED***.To(MatchJQ(".ram", 137438953472.0***REMOVED******REMOVED***

		// Check the AWS machine type:
		awsTypes, err := JQ(`.attributes.items[] | select(.cloud_provider == "aws"***REMOVED***`, resource***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		Expect(awsTypes***REMOVED***.To(HaveLen(1***REMOVED******REMOVED***
		awsType := awsTypes[0]
		Expect(awsType***REMOVED***.To(MatchJQ(".cloud_provider", "aws"***REMOVED******REMOVED***
		Expect(awsType***REMOVED***.To(MatchJQ(".id", "c5.12xlarge"***REMOVED******REMOVED***
		Expect(awsType***REMOVED***.To(MatchJQ(".name", "c5.12xlarge - Compute optimized"***REMOVED******REMOVED***
		Expect(awsType***REMOVED***.To(MatchJQ(".cpu", 48.0***REMOVED******REMOVED***
		Expect(awsType***REMOVED***.To(MatchJQ(".ram", 103079215104.0***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
