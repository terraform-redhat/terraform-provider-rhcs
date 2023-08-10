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

	. "github.com/onsi/ginkgo/v2/dsl/core"
***REMOVED***
	. "github.com/onsi/gomega/ghttp"
	. "github.com/openshift-online/ocm-sdk-go/testing"
***REMOVED***

var _ = Describe("DNS Domain creation", func(***REMOVED*** {
	domain := "my.domain.openshift.dev"

	Context("Verify success", func(***REMOVED*** {
		It("Should create a DNS domain", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				// first post (create***REMOVED***
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/dns_domains",
					***REMOVED***,
					VerifyJSON(`{
                        "kind": "DNSDomain"
                    }`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
	    			  "id": "`+domain+`"
	    	***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			# (resource arguments***REMOVED***
	    ***REMOVED***
	    	`***REMOVED***

			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			resource := terraform.Resource("rhcs_dns_domain", "dns"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", domain***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Should recreate a DNS domain on 404 (reconcile***REMOVED***", func(***REMOVED*** {
			newDomain := "new." + domain
			// Prepare the server for the firs create
			server.AppendHandlers(
				// first post (create***REMOVED***
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/dns_domains",
					***REMOVED***,
					VerifyJSON(`{
                        "kind": "DNSDomain"
                    }`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
	    			  "id": "`+domain+`"
	    	***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			# (resource arguments***REMOVED***
	    ***REMOVED***
	    	`***REMOVED***

			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			resource := terraform.Resource("rhcs_dns_domain", "dns"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", domain***REMOVED******REMOVED***

			// prepare server for the reconcile

			server.AppendHandlers(
				// first is read to update state. lets return 404
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/dns_domains/"+domain,
					***REMOVED***,
					RespondWithJSON(http.StatusNotFound, `{}`***REMOVED***,
				***REMOVED***,
				// Now tf should create a new dns
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/dns_domains",
					***REMOVED***,
					VerifyJSON(`{
                        "kind": "DNSDomain"
                    }`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+newDomain+`",
	    			  "id": "`+newDomain+`"
	    	***REMOVED***`***REMOVED***,
				***REMOVED***,
				// Read the domain to load the current state:
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/dns_domains/"+newDomain,
					***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "kind": "DNSDomain",
			    	  "href": "/api/clusters_mgmt/v1/dns_domains/`+newDomain+`",
			    	  "id": "`+newDomain+`"
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// run terraform

			terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			# (resource arguments***REMOVED***
	    ***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			resource = terraform.Resource("rhcs_dns_domain", "dns"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", newDomain***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("DNS domain import", func(***REMOVED*** {
	domain := "my.domain.openshift.dev"
	It("should import successfully", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// first is for the import state callback
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/dns_domains/"+domain,
				***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "kind": "DNSDomain",
				  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
				  "id": "`+domain+`"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Read the domain to load the current state:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/dns_domains/"+domain,
				***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "kind": "DNSDomain",
				  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
				  "id": "`+domain+`"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
			resource "rhcs_dns_domain" "dns" {
				# (resource arguments***REMOVED***
	***REMOVED***
		`***REMOVED***

		Expect(terraform.Import("rhcs_dns_domain.dns", domain***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_dns_domain", "dns"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", domain***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
