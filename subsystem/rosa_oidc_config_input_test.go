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
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
***REMOVED***             // nolint
***REMOVED***

var _ = Describe("Cluster creation", func(***REMOVED*** {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.

	It("Create oidc config input resource", func(***REMOVED*** {
		terraform.Source(`
				resource "ocm_rosa_oidc_config_input" "oidc_input" {
  					region = "us-east-1"
		***REMOVED***
			`***REMOVED***

		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		// when calling again to apply it should work
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
