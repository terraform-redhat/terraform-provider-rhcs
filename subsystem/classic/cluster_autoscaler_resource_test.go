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

var _ = Describe("Cluster Autoscaler", func() {
	Context("creation", func() {

		It("fails if cluster ID is empty", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = ""
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute cluster cluster ID may not be empty/blank")
		})

		It("fails if given an out-of-range utilization threshold", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					scale_down = {
						utilization_threshold = "1.1"
					}
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Value '1.100000' is out of range 0.000000 - 1.000000")
		})

		It("fails if given an invalid range", func() {
			Terraform.Source(`
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
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("In 'resource_limits.cores' attribute, max value must be greater or equal to min value")
		})

		It("fails if given an invalid duration string", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					max_node_provision_time = "1"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`Value '1' cannot be parsed to a duration string`)
		})

		It("fails to find a matching cluster object", func() {
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
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Cluster '123' not found")
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
					RespondWithJSON(http.StatusInternalServerError, "{}"),
				),
			)

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Failed creating autoscaler for cluster '123': status is 500")
		})

		It("successfully creates a cluster-autoscaler object", func() {
			TestServer.AppendHandlers(
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

			Terraform.Source(`
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

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("validation failure - negative log_verbosity", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					log_verbosity = -3
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute log_verbosity value must be at least 0, got: -3")
		})
		It("validation failure - negative core.min", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					resource_limits = {
						max_nodes_total = 20
						cores = {
							min = -2
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
						]
					}
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute 'resource_limits.cores.min' value must be at least 0, got: -2")
		})
		It("validation failure - negative memory.min", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					resource_limits = {
						max_nodes_total = 20
						cores = {
							min = 0
							max = 1
						}
						memory = {
							min = -3
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
						]
					}
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute 'resource_limits.memory.min' value must be at least 0, got: -3")
		})
		It("validation failure - negative gpu.min", func() {
			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
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
									min = -1
									max = 1
								}
							},
						]
					}
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute 'resource_limits.gpus.[0].range.min' value must be at least 0, got: -1")
		})
	})

	Context("importing", func() {
		It("fails if resource does not exist in OCM", func() {
			TestServer.AppendHandlers(
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

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			runOutput := Terraform.Import("rhcs_cluster_autoscaler.cluster_autoscaler", "123")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Cannot import non-existent remote object")
		})

		It("succeeds if resource exists in OCM", func() {
			TestServer.AppendHandlers(
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

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
				}
	    	`)
			runOutput := Terraform.Import("rhcs_cluster_autoscaler.cluster_autoscaler", "123")
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_cluster_autoscaler", "cluster_autoscaler").(map[string]interface{})
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
			TestServer.AppendHandlers(
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

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
					skip_nodes_with_local_storage = false

				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
		It("Test setting `skip_nodes_with_local_storage` to null ", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true,
							"skip_nodes_with_local_storage": false
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true,
							"skip_nodes_with_local_storage": false
						}
					`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/autoscaler"),
					VerifyJQ(".balance_similar_node_groups", true),
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
						}
					`),
				),
			)

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("successfully applies the changes in OCM", func() {
			TestServer.AppendHandlers(
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

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
					skip_nodes_with_local_storage = true
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_cluster_autoscaler", "cluster_autoscaler").(map[string]interface{})
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
			TestServer.AppendHandlers(
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

			Terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
				}
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("trivially succeeds if the autoscaler object does not exist in OCM", func() {
			TestServer.AppendHandlers(
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

			Expect(Terraform.Destroy().ExitCode).To(BeZero())
		})

		It("successfully applies the deletion in OCM", func() {
			TestServer.AppendHandlers(
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

			Expect(Terraform.Destroy().ExitCode).To(BeZero())
		})

	})
})
