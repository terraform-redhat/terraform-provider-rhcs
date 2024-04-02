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

package provider

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"                      // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Cluster KubeletConfig", func() {
	Context("Create KubeletConfig", func() {
		It("fails if cluster ID is empty", func() {
			terraform.Source(`
			resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = ""
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})
		It("It fails if the requested podPidsLimit is below the minimum", func() {
			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 4000
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("It fails if the requested podPidsLimit is above the unsafe maximum", func() {
			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 4000000
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("It fails if the cluster does not exist", func() {
			server.AppendHandlers(
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

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("fails if OCM backend fails to create the object", func() {
			server.AppendHandlers(
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusInternalServerError, "Internal Server Error"),
				),
			)

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Successfully creates a KubeletConfig", func() {
			server.AppendHandlers(
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})
	})

	Context("importing", func() {
		It("fails if resource does not exist in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
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

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Import("rhcs_kubeletconfig.cluster_kubeletconfig", "123")).ToNot(BeZero())
		})

		It("succeeds if resource exists in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config",
							"pod_pids_limit": 5000
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubelet_config" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Import("rhcs_kubeletconfig.cluster_kubelet_config", "123")).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_kubeletconfig", "cluster_kubelet_config").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":        "123",
					"pod_pids_limit": float64(5000),
				},
			))
		})
	})

	Context("updating", func() {
		BeforeEach(func() {
			server.AppendHandlers(
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("successfully applies the changes in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					VerifyJQ(".pod_pids_limit", float64(10000)),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config",
							"pod_pids_limit": 10000
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 10000
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_kubeletconfig", "cluster_kubeletconfig").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":        "123",
					"pod_pids_limit": float64(10000),
				},
			))
		})
	})

	Context("deletion", func() {
		BeforeEach(func() {
			server.AppendHandlers(
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					VerifyJQ(".pod_pids_limit", float64(5000)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "KubeletConfig",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_kubeletconfig" "cluster_kubeletconfig" {
					cluster = "123"
					pod_pids_limit = 5000
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("trivially succeeds if the kubeletconfig object does not exist in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
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

			Expect(terraform.Destroy()).To(BeZero())
		})

		It("successfully applies the deletion in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "KubeletConfig",
							"href": "/api/clusters_mgmt/v1/clusters/123/kubelet_config",
							"pod_pids_limit": 5000
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/kubelet_config"),
					RespondWithJSON(http.StatusNoContent, "{}"),
				),
			)

			Expect(terraform.Destroy()).To(BeZero())
		})
	})
})
