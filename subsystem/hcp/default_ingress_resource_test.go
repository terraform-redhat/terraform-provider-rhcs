/*
Copyright (c) 2024 Red Hat, Inc.

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

	. "github.com/onsi/ginkgo/v2"                      // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_cluster_rosa_hcp - default ingress", func() {

	defaultDay1Template := `{
						 "kind": "IngressList",
						 "href": "/api/clusters_mgmt/v1/clusters/123/ingresses",
						 "page": 1,
						 "size": 1,
						 "total": 1,
						 "items": [
						   {
							 "kind": "Ingress",
							 "href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
							 "id": "d6z2",
							 "listening": "external",
							 "default": true,
							 "dns_name": "redhat.com"
						   }
						 ]
						}
						`
	clusterReady := `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`

	It("fails if listening method is not supplied", func() {
		Terraform.Source(`
			resource "rhcs_hcp_default_ingress" "default_ingress" {
				cluster = ""
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring(`The argument "listening_method" is required`)
	})

	It("fails if cluster ID is empty", func() {
		Terraform.Source(`
			resource "rhcs_hcp_default_ingress" "default_ingress" {
				listening_method = "internal"
				cluster = ""
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("Attribute cluster cluster ID may not be empty/blank string")
	})

	It("Updates ListeningMethod", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, clusterReady),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses"),
				RespondWithJSON(http.StatusOK, defaultDay1Template),
			),

			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				VerifyJQ(".listening", "internal"),
				RespondWithJSON(http.StatusOK, `
				{
					"kind": "Ingress",
					"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
					"id": "d6z2",
					"listening": "internal",
					"default": true,
					"dns_name": "redhat.com"
				}`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
			resource "rhcs_hcp_default_ingress" "default_ingress" {
			cluster = "123"
			listening_method = "internal"
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Update default ingress and delete it", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, clusterReady),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses"),
				RespondWithJSON(http.StatusOK, defaultDay1Template),
			),

			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				VerifyJQ(".listening", "internal"),
				RespondWithJSON(http.StatusOK, `
						 {
							 "kind": "Ingress",
							 "href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
							 "id": "d6z2",
							 "listening": "internal",
							 "default": true,
							 "dns_name": "redhat.com"
						}
						`),
			),
		)

		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_hcp_default_ingress" "default_ingress" {
			cluster = "123"
		    listening_method = "internal"
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				RespondWithJSON(http.StatusOK, `
						 {
							 "kind": "Ingress",
							 "href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
							 "id": "d6z2",
							 "listening": "internal",
							 "default": true,
							 "dns_name": "redhat.com"
						}
						`),
			),
		)

		// remove ingress
		Terraform.Source("")
		// Last pool, we ignore the error, so this succeeds
		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

	})

})
