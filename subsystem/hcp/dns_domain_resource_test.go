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

package hcp

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing"
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("DNS Domain creation", func() {
	domain := "my.domain.openshift.dev"

	Context("Verify success", func() {
		When("cluster arch is specified sets state", func() {
			It("Should create a DNS domain", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					// first post (create)
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/dns_domains",
						),
						VerifyJSON(`{
                        "kind": "DNSDomain",
						"cluster_arch": "hcp"
                    }`),
						RespondWithJSON(http.StatusOK, `{
	    			  "kind": "DNSDomain",
	    			  "href": "/api/clusters_mgmt/v1/dns_domains/`+domain+`",
	    			  "id": "`+domain+`",
					  "cluster_arch": "`+string(v1.ClusterArchitectureHcp)+`"
	    			}`),
					),
				)

				Terraform.Source(`
	    		resource "rhcs_dns_domain" "dns" {
	    			cluster_arch = "hcp"
	    		}
	    	`)

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				resource := Terraform.Resource("rhcs_dns_domain", "dns")
				Expect(resource).To(MatchJQ(".attributes.id", domain))
				Expect(resource).To(MatchJQ(".attributes.cluster_arch", string(v1.ClusterArchitectureHcp)))
			})
		})
	})
})
