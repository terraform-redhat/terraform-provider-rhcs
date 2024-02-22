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
)

var _ = Describe("Hcp Cluster Autoscaler", func() {
	Context("creation", func() {
		It("fails if given an invalid duration string", func() {
			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					max_node_provision_time = "1"
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("fails to find a matching cluster object", func() {
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
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusInternalServerError, "Internal Server Error"),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("successfully creates a cluster-autoscaler object", func() {
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".max_pod_grace_period", float64(1)),
					VerifyJQ(".pod_priority_threshold", float64(-10)),
					VerifyJQ(".max_node_provision_time", "1h"),
					VerifyJQ(".resource_limits.max_nodes_total", float64(20)),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "ClusterAutoscaler",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					max_pod_grace_period = 1
					pod_priority_threshold = -10
					max_node_provision_time = "1h"
					resource_limits = {
						max_nodes_total = 20
					}
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})
	})

	Context("importing", func() {
		It("fails if resource does not exist in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Autoscaler for cluster ID '123' is not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Import("rhcs_hcp_cluster_autoscaler.cluster_autoscaler", "123")).ToNot(BeZero())
		})

		It("succeeds if resource exists in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"max_node_provision_time": "1h"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"max_node_provision_time": "1h"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Import("rhcs_hcp_cluster_autoscaler.cluster_autoscaler", "123")).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_hcp_cluster_autoscaler", "cluster_autoscaler").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":                 "123",
					"max_pod_grace_period":    nil,
					"pod_priority_threshold":  nil,
					"max_node_provision_time": "1h",
					"resource_limits":         nil,
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".max_node_provision_time", "1h"),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "ClusterAutoscaler",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					max_node_provision_time = "1h"

				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("successfully applies the changes in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".max_node_provision_time", "2h"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"max_node_provision_time": "2h"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					max_node_provision_time = "2h"
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_hcp_cluster_autoscaler", "cluster_autoscaler").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":                 "123",
					"max_pod_grace_period":    nil,
					"pod_priority_threshold":  nil,
					"max_node_provision_time": "2h",
					"resource_limits":         nil,
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "ClusterAutoscaler",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_hcp_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("trivially succeeds if the autoscaler object does not exist in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Autoscaler for cluster ID '123' is not found",
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
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusNoContent, "{}"),
				),
			)

			Expect(terraform.Destroy()).To(BeZero())
		})

	})
})
