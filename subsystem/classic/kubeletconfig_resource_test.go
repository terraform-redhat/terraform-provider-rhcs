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

var _ = Describe("Cluster KubeletConfig", func() {
	Context("Create KubeletConfig", func() {
		It("fails if pod_pids_limit is empty", func() {
			Terraform.Source(`
			resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = ""
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`The argument "pod_pids_limit" is required`)
		})
		It("fails if cluster is empty", func() {
			Terraform.Source(`
			resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					pod_pids_limit = 5000
					cluster = ""
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`Attribute cluster cluster ID may not be empty/blank string`)
		})
		It("It fails if the requested podPidsLimit is below the minimum", func() {
			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 4000
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("The requested podPidsLimit of '4000' is below the minimum allowable")
		})

		It("It fails if the requested podPidsLimit is above the unsafe maximum", func() {
			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 4000000
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("The requested podPidsLimit of '4000000' is above the default maximum")
		})

		It("It fails if the cluster does not exist", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Cluster '123' not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Cluster '123' not found")
		})

		It("fails if there is already one kubelet config for classic cluster", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"hypershift": {
								"enabled": false
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, `
						{
							"items": [
								{
								  "kind": "KubeletConfig",
								  "id": "456",
								  "href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
								  "name": "my_name",
								  "pod_pids_limit": 5000
								}
							  ]
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusInternalServerError, "Internal Server Error"),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("KubeletConfig for cluster '123' already exist")
		})

		It("Successfully creates two KubeletConfigs for HCP cluster", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready",
							"hypershift": {
								"enabled": true
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"name": "my_name",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"name": "my_name",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready",
							"hypershift": {
								"enabled": true
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					VerifyJQ(".pod_pids_limit", float64(10000)),
					VerifyJQ(".name", "custom_name"),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "789",
							"name": "custom_name",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/789",
							"pod_pids_limit": 10000
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}

				resource "rhcs_kubeletconfig" "cluster_kubeletconfig_1" {
					cluster = "123"
					pod_pids_limit = 10000
					name = "custom_name"
				}
	    	`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("fails if OCM backend fails to create the object", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"hypershift": {
								"enabled": false
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusInternalServerError, "{}"),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Failed to create KubeletConfig on cluster '123': status is 500")
		})

		It("Successfully creates a KubeletConfig", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready",
							"hypershift": {
								"enabled": false
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"name": "my_name",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"pod_pids_limit": 5000
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("importing", func() {
		It("fails if resource does not exist in OCM", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
				}
	    	`)
			runOutput := Terraform.Import("rhcs_kubeletconfig.cluster_kubeletconfig", "123")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Cannot find KubeletConfig for cluster '123'")
		})

		It("succeeds if resource exists in OCM", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, `
						{
							"items": [
								{
								  "kind": "KubeletConfig",
								  "id": "456",
								  "href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
								  "name": "my_name",
								  "pod_pids_limit": 5000
								}
							  ]
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"name": "my_name",
							"pod_pids_limit": 5000
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubelet_config" {
					cluster = "123"
				}
	    	`)
			runOutput := Terraform.Import("rhcs_kubeletconfig.cluster_kubelet_config", "123")
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_kubeletconfig", "cluster_kubelet_config").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":        "123",
					"id":             "456",
					"name":           "my_name",
					"pod_pids_limit": float64(5000),
				},
			))
		})
	})

	Context("updating", func() {
		BeforeEach(func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready",
							"hypershift": {
								"enabled": false
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"name": "my_name",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"pod_pids_limit": 5000
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("successfully applies the changes in OCM", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"name": "my_name",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"name": "my_name",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					VerifyJQ(".pod_pids_limit", float64(10000)),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"name": "my_name",
							"pod_pids_limit": 10000
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 10000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_kubeletconfig", "cluster_kubeletconfig").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":        "123",
					"pod_pids_limit": float64(10000),
					"id":             "456",
					"name":           "my_name",
				},
			))
		})
	})

	Context("deletion", func() {
		BeforeEach(func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready",
							"hypershift": {
								"enabled": false
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"name": "my_name",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"pod_pids_limit": 5000
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("trivially succeeds if the kubeletconfig object does not exist in OCM", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "KubeletConfig for cluster ID '123' is not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
						}
					`),
				),
			)

			Expect(Terraform.Destroy().ExitCode).To(BeZero())
		})

		It("successfully applies the deletion in OCM", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"id": "456",
							"name": "my_name",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/kubelet_configs/456"),
					RespondWithJSON(http.StatusNoContent, "{}"),
				),
			)

			Expect(Terraform.Destroy().ExitCode).To(BeZero())
		})
	})
})
