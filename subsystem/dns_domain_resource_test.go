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

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/openshift-online/ocm-sdk-go/testing"
)

var _ = Describe("DNS Domain creation", func() {
	domain := "my.domain.openshift.dev"

	Context("Verify success", func() {
		It("Should create a DNS domain", func() {
			// Prepare the server:
			server.AppendHandlers(
				// first post (create)
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/dns_domains",
					),
					VerifyJSON(`{
                        "kind": "DNSDomain"
                    }`),
					RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
	    			  "id": "`+domain+`"
	    			}`),
				),
			)

			terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			# (resource arguments)
	    		}
	    	`)

			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_dns_domain", "dns")
			Expect(resource).To(MatchJQ(".attributes.id", domain))
		})

		It("Should recreate a DNS domain on 404 (reconcile)", func() {
			newDomain := "new." + domain
			// Prepare the server for the firs create
			server.AppendHandlers(
				// first post (create)
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/dns_domains",
					),
					VerifyJSON(`{
                        "kind": "DNSDomain"
                    }`),
					RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
	    			  "id": "`+domain+`"
	    			}`),
				),
			)

			terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			# (resource arguments)
	    		}
	    	`)

			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_dns_domain", "dns")
			Expect(resource).To(MatchJQ(".attributes.id", domain))

			// prepare server for the reconcile

			server.AppendHandlers(
				// first is read to update state. lets return 404
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/dns_domains/"+domain,
					),
					RespondWithJSON(http.StatusNotFound, `{}`),
				),
				// Now tf should create a new dns
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/dns_domains",
					),
					VerifyJSON(`{
                        "kind": "DNSDomain"
                    }`),
					RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+newDomain+`",
	    			  "id": "`+newDomain+`"
	    			}`),
				),
				// Read the domain to load the current state:
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/dns_domains/"+newDomain,
					),
					RespondWithJSON(http.StatusOK, `{
			    	  "kind": "DNSDomain",
			    	  "href": "/api/clusters_mgmt/v1/dns_domains/`+newDomain+`",
			    	  "id": "`+newDomain+`"
			    	}`),
				),
			)

			// run terraform

			terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			# (resource arguments)
	    		}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_dns_domain", "dns")
			Expect(resource).To(MatchJQ(".attributes.id", newDomain))
		})
	})
})

var _ = Describe("DNS domain import", func() {
	domain := "my.domain.openshift.dev"
	It("should import successfully", func() {
		// Prepare the server:
		server.AppendHandlers(
			// first is for the import state callback
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/dns_domains/"+domain,
				),
				RespondWithJSON(http.StatusOK, `{
				  "kind": "DNSDomain",
				  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
				  "id": "`+domain+`"
				}`),
			),
			// Read the domain to load the current state:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/dns_domains/"+domain,
				),
				RespondWithJSON(http.StatusOK, `{
				  "kind": "DNSDomain",
				  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
				  "id": "`+domain+`"
				}`),
			),
		)

		terraform.Source(`
			resource "rhcs_dns_domain" "dns" {
				# (resource arguments)
			}
		`)

		Expect(terraform.Import("rhcs_dns_domain.dns", domain)).To(BeZero())
		resource := terraform.Resource("rhcs_dns_domain", "dns")
		Expect(resource).To(MatchJQ(".attributes.id", domain))
	})
})
