/*
Copyright (c) 2023 Red Hat, Inc.

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

package classic

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"                      // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("default ingress", func() {

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
							 "dns_name": "redhat.com",
							 "load_balancer_type": "classic",
							 "route_wildcard_policy": "WildcardsDisallowed",
							 "route_namespace_ownership_policy": "Strict"
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

	It("fails if cluster ID is empty", func() {
		Terraform.Source(`
			resource "rhcs_default_ingress" "default_ingress" {
				cluster = ""
			}
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("Attribute cluster cluster ID may not be empty/blank string")
	})

	It("Sends updates to attribute individually", func() {
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
				VerifyJQ(`.route_selectors`, map[string]interface{}{}),
				VerifyJQ(`.excluded_namespaces`, []interface{}{"stage", "int", "aaa"}),
				VerifyJQ(`.load_balancer_type`, "nlb"),
				RespondWithJSON(http.StatusOK, `
				{
					"kind": "Ingress",
					"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
					"id": "d6z2",
					"listening": "external",
					"default": true,
					"dns_name": "redhat.com",
					"load_balancer_type": "nlb",
					"excluded_namespaces": [
						"stage",
						"int",
						"aaa"
					],
					"route_wildcard_policy": "WildcardsDisallowed",
					"route_namespace_ownership_policy": "Strict"
				}`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
			resource "rhcs_default_ingress" "default_ingress" {
			id = "d6z2"
			cluster = "123"
			excluded_namespaces = ["stage", "int", "aaa"]
			load_balancer_type = "nlb"
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
					"listening": "external",
					"default": true,
					"dns_name": "redhat.com",
					"load_balancer_type": "nlb",
					"excluded_namespaces": [
						"stage",
						"int",
						"aaa"
					],
					"route_wildcard_policy": "WildcardsDisallowed",
					"route_namespace_ownership_policy": "Strict"
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				VerifyJQ(`.load_balancer_type`, nil),
				VerifyJQ(`.route_selectors`, nil),
				VerifyJQ(`.route_wildcard_policy`, nil),
				VerifyJQ(`.route_namespace_ownership_policy`, nil),
				VerifyJQ(`.excluded_namespaces`, []interface{}{"int", "aaa"}),
				RespondWithJSON(http.StatusOK, `
				{
					"kind": "Ingress",
					"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
					"id": "d6z2",
					"listening": "external",
					"default": true,
					"dns_name": "redhat.com",
					"load_balancer_type": "nlb",
					"excluded_namespaces": [
						"int",
						"aaa"
					],
					"route_wildcard_policy": "WildcardsDisallowed",
					"route_namespace_ownership_policy": "Strict"
				}`),
			),
		)

		// Run the apply command:
		Terraform.Source(`
			resource "rhcs_default_ingress" "default_ingress" {
			id = "d6z2"
			cluster = "123"
			excluded_namespaces = ["int", "aaa"]
			load_balancer_type = "nlb"
		}`)
		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				RespondWithJSON(http.StatusOK, `
				{
					"kind": "Ingress",
					"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
					"id": "d6z2",
					"listening": "external",
					"default": true,
					"dns_name": "redhat.com",
					"load_balancer_type": "nlb",
					"excluded_namespaces": [
						"int",
						"aaa"
					],
					"route_wildcard_policy": "WildcardsDisallowed",
					"route_namespace_ownership_policy": "Strict"
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				VerifyJQ(`.load_balancer_type`, "classic"),
				VerifyJQ(`.route_selectors`, nil),
				VerifyJQ(`.route_wildcard_policy`, nil),
				VerifyJQ(`.route_namespace_ownership_policy`, nil),
				VerifyJQ(`.excluded_namespaces`, nil),
				RespondWithJSON(http.StatusOK, `
				{
					"kind": "Ingress",
					"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
					"id": "d6z2",
					"listening": "external",
					"default": true,
					"dns_name": "redhat.com",
					"load_balancer_type": "classic",
					"excluded_namespaces": [
						"int",
						"aaa"
					],
					"route_wildcard_policy": "WildcardsDisallowed",
					"route_namespace_ownership_policy": "Strict"
				}`),
			),
		)

		// Run the apply command:
		Terraform.Source(`
			resource "rhcs_default_ingress" "default_ingress" {
			id = "d6z2"
			cluster = "123"
			excluded_namespaces = ["int", "aaa"]
			load_balancer_type = "classic"
		}`)
		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default ingress - excluded_namespaces and load balancer type set", func() {
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
				VerifyJQ(`.route_selectors`, map[string]interface{}{}),
				VerifyJQ(`.excluded_namespaces`, []interface{}{"stage", "int", "aaa"}),
				VerifyJQ(`.load_balancer_type`, "nlb"),
				RespondWithJSON(http.StatusOK, `
						 {
							 "kind": "Ingress",
							 "href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
							 "id": "d6z2",
							 "listening": "external",
							 "default": true,
							 "dns_name": "redhat.com",
							 "load_balancer_type": "nlb",
							 "excluded_namespaces": [
								"stage",
								"int",
								"aaa"
							 ],
							 "route_wildcard_policy": "WildcardsDisallowed",
							 "route_namespace_ownership_policy": "Strict"
						}
						`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
		    excluded_namespaces = ["stage", "int", "aaa"]
			load_balancer_type = "nlb"
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default -  failed if only cluster_routes_tls_secret_ref set", func() {
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
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123",
			cluster_routes_tls_secret_ref = "111"
		}`)
		Expect(Terraform.Apply()).NotTo(BeZero())
	})

	It("Create cluster with default ingress - routers_selectors set, route_wildcard_policy and InterNamespaceAllowed changed", func() {
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
				VerifyJQ(`.route_selectors`, map[string]interface{}{"foo": "bar"}),
				VerifyJQ(`.excluded_namespaces`, []interface{}{}),
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_selectors": {
							"foo": "bar"
						},
						"route_wildcard_policy": "WildcardsAllowed",
						"route_namespace_ownership_policy": "InterNamespaceAllowed"
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        route_wildcard_policy = "WildcardsAllowed"
			route_selectors = {
			   "foo" = "bar",
			}
			route_namespace_ownership_policy = "InterNamespaceAllowed"
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("routers_selectors and external namespaces cleanup", func() {
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
				VerifyJQ(`.route_selectors`, map[string]interface{}{"foo": "bar"}),
				VerifyJQ(`.excluded_namespaces`, []interface{}{"stage", "int", "aaa"}),
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_selectors": {
							"foo": "bar"
						},
						"excluded_namespaces": [
							"stage",
							"int",
							"aaa"
						],
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict"
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
			route_selectors = {
			   "foo" = "bar",
			}
			excluded_namespaces = ["stage", "int", "aaa"]
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
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_selectors": {
							"foo": "bar"
						},
						"excluded_namespaces": [
							"stage",
							"int",
							"aaa"
						],
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict"
					}
				`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				VerifyJQ(`.route_selectors`, map[string]interface{}{}),
				VerifyJQ(`.excluded_namespaces`, []interface{}{}),
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict"
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
		}`)
		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default ingress - cluster_routes_hostname and cluster_routes_tls_secret_ref set to actual value", func() {
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
				VerifyJQ(`.cluster_routes_tls_secret_ref`, "111"),
				VerifyJQ(`.cluster_routes_hostname`, "aaa"),
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"cluster_routes_tls_secret_ref": "111",
						"cluster_routes_hostname": "aaa",
						"route_namespace_ownership_policy": "Strict"
						
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        cluster_routes_tls_secret_ref = "111"
			cluster_routes_hostname =  "aaa"
			
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default ingress - cluster_routes_hostname set to empty value", func() {
		// Run the apply command:
		Terraform.Source(`resource "rhcs_default_ingress" "default_ingress" {
				cluster = "123"
				cluster_routes_hostname =  ""
			}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("Attribute cluster_routes_hostname string length must be at least 1")
	})

	It("Create cluster with default ingress - cluster_routes_tls_secret_ref set to empty value", func() {
		// Run the apply command:
		Terraform.Source(`resource "rhcs_default_ingress" "default_ingress" {
				cluster = "123"
				cluster_routes_tls_secret_ref = ""
			}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("Attribute cluster_routes_tls_secret_ref string length must be at least 1")
	})

	It("Create cluster with default ingress - component_routes set fully", func() {
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
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"console": {
							  "hostname": "console-host",
							  "tls_secret_ref": "console-secret"
							},
							"downloads": {
							  "hostname": "downloads-host",
							  "tls_secret_ref": "downloads-secret"
							},
							"oauth": {
							  "hostname": "oauth-host-new",
							  "tls_secret_ref": "oauth-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        component_routes = {
				"oauth" = {
				  "hostname"       = "oauth-host-new"
				  "tls_secret_ref" = "oauth-secret"
				}
				"console" = {
				  "hostname"       = "console-host"
				  "tls_secret_ref" = "console-secret"
				}
				"downloads" = {
				  "hostname"       = "downloads-host"
				  "tls_secret_ref" = "downloads-secret"
				}
			}
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default ingress - component_routes set single attribute", func() {
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
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"console": {
							  "hostname": "console-host",
							  "tls_secret_ref": "console-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        component_routes = {
				"console" = {
				  "hostname"       = "console-host"
				  "tls_secret_ref" = "console-secret"
				}
			}
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default ingress - component_routes fails to set single attribute as nil", func() {
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
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"downloads": {
							  "hostname": "downloads-host",
							  "tls_secret_ref": "downloads-secret"
							},
							"oauth": {
							  "hostname": "oauth-host-new",
							  "tls_secret_ref": "oauth-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        component_routes = {
				"oauth" = {
				  "hostname"       = "oauth-host-new"
				  "tls_secret_ref" = "oauth-secret"
				}
				"console" = null
				"downloads" = {
				  "hostname"       = "downloads-host"
				  "tls_secret_ref" = "downloads-secret"
				}
			}
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("Component route shouldn't be null, if you would like to reset a specific component route please remove the key instead")
	})

	It("Create cluster with default ingress - component_routes fails to set single route as empty", func() {
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
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"downloads": {
							  "hostname": "downloads-host",
							  "tls_secret_ref": "downloads-secret"
							},
							"oauth": {
							  "hostname": "oauth-host-new",
							  "tls_secret_ref": "oauth-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        component_routes = {
				"oauth" = {
				  "hostname"       = "oauth-host-new"
				  "tls_secret_ref" = "oauth-secret"
				}
				"console" = {
				  "hostname"       = ""
				  "tls_secret_ref" = ""
				}
				"downloads" = {
				  "hostname"       = "downloads-host"
				  "tls_secret_ref" = "downloads-secret"
				}
			}
		}`)
		runOutput := Terraform.Apply()
		runOutput.VerifyErrorContainsSubstring("Component route fields shouldn't both be empty, if you would like to reset a specific component route please remove the key instead")
	})

	It("Create cluster with default ingress - component_routes reset single attribute", func() {
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
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"console": {
							  "hostname": "console-host",
							  "tls_secret_ref": "console-secret"
							},
							"downloads": {
							  "hostname": "downloads-host",
							  "tls_secret_ref": "downloads-secret"
							},
							"oauth": {
							  "hostname": "oauth-host-new",
							  "tls_secret_ref": "oauth-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        component_routes = {
				"oauth" = {
				  "hostname"       = "oauth-host-new"
				  "tls_secret_ref" = "oauth-secret"
				}
				"console" = {
				  "hostname"       = "console-host"
				  "tls_secret_ref" = "console-secret"
				}
				"downloads" = {
				  "hostname"       = "downloads-host"
				  "tls_secret_ref" = "downloads-secret"
				}
			}
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				RespondWithJSON(http.StatusOK, defaultDay1Template),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses"),
				RespondWithJSON(http.StatusOK, defaultDay1Template),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				VerifyJSON(`
				{
        			"kind": "Ingress",
        			"component_routes": {
          				"console": {
            				"kind": "ComponentRoute",
            				"hostname": "console-host",
            				"tls_secret_ref": "console-secret"
          				},
          				"downloads": {
				            "kind": "ComponentRoute",
            				"hostname": "downloads-host",
            				"tls_secret_ref": "downloads-secret"
          				},
          				"oauth": {
							"kind": "ComponentRoute",
            				"hostname": "",
            				"tls_secret_ref": ""
          				}
        			}
      			}`),
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"console": {
							  "hostname": "console-host",
							  "tls_secret_ref": "console-secret"
							},
							"downloads": {
							  "hostname": "downloads-host",
							  "tls_secret_ref": "downloads-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
			id = "d6z2"
	        component_routes = {
				"console" = {
				  "hostname"       = "console-host"
				  "tls_secret_ref" = "console-secret"
				}
				"downloads" = {
				  "hostname"       = "downloads-host"
				  "tls_secret_ref" = "downloads-secret"
				}
			}
		}`)
		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create cluster with default ingress - component_routes reset fully", func() {
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
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict",
						"component_routes": {
							"console": {
							  "hostname": "console-host",
							  "tls_secret_ref": "console-secret"
							},
							"downloads": {
							  "hostname": "downloads-host",
							  "tls_secret_ref": "downloads-secret"
							},
							"oauth": {
							  "hostname": "oauth-host-new",
							  "tls_secret_ref": "oauth-secret"
							}
						}
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
	        component_routes = {
				"oauth" = {
				  "hostname"       = "oauth-host-new"
				  "tls_secret_ref" = "oauth-secret"
				}
				"console" = {
				  "hostname"       = "console-host"
				  "tls_secret_ref" = "console-secret"
				}
				"downloads" = {
				  "hostname"       = "downloads-host"
				  "tls_secret_ref" = "downloads-secret"
				}
			}
		}`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				RespondWithJSON(http.StatusOK, defaultDay1Template),
			),

			CombineHandlers(
				VerifyJSON(`
				{
        			"kind": "Ingress",
        			"component_routes": {
          				"console": {
            				"kind": "ComponentRoute",
            				"hostname": "",
            				"tls_secret_ref": ""
          				},
          				"downloads": {
				            "kind": "ComponentRoute",
            				"hostname": "",
            				"tls_secret_ref": ""
          				},
          				"oauth": {
							"kind": "ComponentRoute",
            				"hostname": "",
            				"tls_secret_ref": ""
          				}
        			},
				"excluded_namespaces": [],
				"route_selectors": {}
      			}`),
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2"),
				RespondWithJSON(http.StatusOK, `
					{
						"kind": "Ingress",
						"href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
						"id": "d6z2",
						"listening": "external",
						"default": true,
						"dns_name": "redhat.com",
						"load_balancer_type": "classic",
						"route_wildcard_policy": "WildcardsDisallowed",
						"route_namespace_ownership_policy": "Strict"
					}
				`),
			),
		)
		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
		}`)
		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("Create default ingress and delete it", func() {
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
				VerifyJQ(`.route_selectors`, map[string]interface{}{}),
				VerifyJQ(`.excluded_namespaces`, []interface{}{"stage", "int", "aaa"}),
				VerifyJQ(`.load_balancer_type`, "nlb"),
				RespondWithJSON(http.StatusOK, `
						 {
							 "kind": "Ingress",
							 "href": "/api/clusters_mgmt/v1/clusters/123/ingresses/d6z2",
							 "id": "d6z2",
							 "listening": "external",
							 "default": true,
							 "dns_name": "redhat.com",
							 "load_balancer_type": "nlb",
							 "excluded_namespaces": [
								"stage",
								"int",
								"aaa"
							 ],
							 "route_wildcard_policy": "WildcardsDisallowed",
							 "route_namespace_ownership_policy": "Strict"
						}
						`),
			),
		)

		// Run the apply command:
		Terraform.Source(`
		  resource "rhcs_default_ingress" "default_ingress" {
			cluster = "123"
		    excluded_namespaces = ["stage", "int", "aaa"]
			load_balancer_type = "nlb"
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
							 "listening": "external",
							 "default": true,
							 "dns_name": "redhat.com",
							 "load_balancer_type": "nlb",
							 "excluded_namespaces": [
								"stage",
								"int",
								"aaa"
							 ],
							 "route_wildcard_policy": "WildcardsDisallowed",
							 "route_namespace_ownership_policy": "Strict"
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
