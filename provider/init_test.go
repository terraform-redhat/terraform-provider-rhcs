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

	. "github.com/onsi/ginkgo" // nolint
***REMOVED*** // nolint
***REMOVED***

var _ = Describe("Init", func(***REMOVED*** {
	var ctx context.Context

	BeforeEach(func(***REMOVED*** {
		ctx = context.Background(***REMOVED***
	}***REMOVED***

	It("Downloads and installs the provider", func(***REMOVED*** {
		result := NewTerraformRunner(***REMOVED***.
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source  = "localhost/openshift-online/ocm"
				    }
				  }
		***REMOVED***
				`,
			***REMOVED***.
			Args("validate"***REMOVED***.
			Run(ctx***REMOVED***
		Expect(result.ExitCode(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
