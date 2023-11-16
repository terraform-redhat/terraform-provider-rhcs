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

var _ = Describe("Cluster Autoscaler", func() {
	Context("creation", func() {
		It("fails if given an out-of-range utilization threshold", func() {
			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					scale_down = {
						utilization_threshold = "1.1"
					}
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("fails if given an invalid range", func() {
			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					resource_limits = {
						cores = {
							min = 1
							max = 0
						}
					}
				}
			`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("fails if given an invalid duration string", func() {
			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
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
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
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
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusInternalServerError, "Internal Server Error"),
				),
			)

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
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
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".balance_similar_node_groups", true),
					VerifyJQ(".skip_nodes_with_local_storage", true),
					VerifyJQ(".log_verbosity", float64(3)),
					VerifyJQ(".max_pod_grace_period", float64(1)),
					VerifyJQ(".pod_priority_threshold", float64(-10)),
					VerifyJQ(".ignore_daemonsets_utilization", false),
					VerifyJQ(".max_node_provision_time", "1h"),
					VerifyJQ(".balancing_ignored_labels", []interface{}{"l1", "l2"}),
					VerifyJQ(".resource_limits.max_nodes_total", float64(20)),
					VerifyJQ(".resource_limits.cores.min", float64(0)),
					VerifyJQ(".resource_limits.cores.max", float64(1)),
					VerifyJQ(".resource_limits.memory.min", float64(2)),
					VerifyJQ(".resource_limits.memory.max", float64(3)),
					VerifyJQ(".resource_limits.gpus[0].type", "nvidia"),
					VerifyJQ(".resource_limits.gpus[0].range.min", float64(0)),
					VerifyJQ(".resource_limits.gpus[0].range.max", float64(1)),
					VerifyJQ(".resource_limits.gpus[1].type", "intel"),
					VerifyJQ(".resource_limits.gpus[1].range.min", float64(2)),
					VerifyJQ(".resource_limits.gpus[1].range.max", float64(3)),
					VerifyJQ(".scale_down.enabled", true),
					VerifyJQ(".scale_down.utilization_threshold", "0.4"),
					VerifyJQ(".scale_down.unneeded_time", "3h"),
					VerifyJQ(".scale_down.delay_after_add", "4h"),
					VerifyJQ(".scale_down.delay_after_delete", "5h"),
					VerifyJQ(".scale_down.delay_after_failure", "6h"),
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
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
					skip_nodes_with_local_storage = true
					log_verbosity = 3
					max_pod_grace_period = 1
					pod_priority_threshold = -10
					ignore_daemonsets_utilization = false
					max_node_provision_time = "1h"
					balancing_ignored_labels = ["l1", "l2"]
					resource_limits = {
						max_nodes_total = 20
						cores = {
							min = 0
							max = 1
						}
						memory = {
							min = 2
							max = 3
						}
						gpus = [
							{
								type = "nvidia"
								range = {
									min = 0
									max = 1
								}
							},
							{
								type = "intel"
								range = {
									min = 2
									max = 3
								}
							}
						]
					}
					scale_down = {
						enabled = true
						utilization_threshold = "0.4"
						unneeded_time = "3h"
						delay_after_add = "4h"
						delay_after_delete = "5h"
						delay_after_failure = "6h"
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
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Import("rhcs_cluster_autoscaler.cluster_autoscaler", "123")).ToNot(BeZero())
		})

		It("succeeds if resource exists in OCM", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"scale_down": {
								"delay_after_add": "2h"
							}
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"scale_down": {
								"delay_after_add": "2h"
							}
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			Expect(terraform.Import("rhcs_cluster_autoscaler.cluster_autoscaler", "123")).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_cluster_autoscaler", "cluster_autoscaler").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":                       "123",
					"balance_similar_node_groups":   nil,
					"skip_nodes_with_local_storage": nil,
					"log_verbosity":                 nil,
					"max_pod_grace_period":          nil,
					"pod_priority_threshold":        nil,
					"ignore_daemonsets_utilization": nil,
					"max_node_provision_time":       nil,
					"balancing_ignored_labels":      nil,
					"resource_limits":               nil,
					"scale_down": map[string]interface{}{
						"enabled":               nil,
						"unneeded_time":         nil,
						"utilization_threshold": nil,
						"delay_after_add":       "2h",
						"delay_after_delete":    nil,
						"delay_after_failure":   nil,
					},
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
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".balance_similar_node_groups", true),
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
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
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
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".balance_similar_node_groups", true),
					VerifyJQ(".skip_nodes_with_local_storage", true),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true,
							"skip_nodes_with_local_storage": true
						}
					`),
				),
			)

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
					skip_nodes_with_local_storage = true
				}
	    	`)
			Expect(terraform.Apply()).To(BeZero())

			actualResource, ok := terraform.Resource("rhcs_cluster_autoscaler", "cluster_autoscaler").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			Expect(actualResource["attributes"]).To(Equal(
				map[string]interface{}{
					"cluster":                       "123",
					"balance_similar_node_groups":   true,
					"skip_nodes_with_local_storage": true,
					"log_verbosity":                 nil,
					"max_pod_grace_period":          nil,
					"pod_priority_threshold":        nil,
					"ignore_daemonsets_utilization": nil,
					"max_node_provision_time":       nil,
					"balancing_ignored_labels":      nil,
					"resource_limits":               nil,
					"scale_down":                    nil,
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
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".balance_similar_node_groups", true),
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
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
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
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
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
