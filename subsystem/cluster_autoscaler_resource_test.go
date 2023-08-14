/*
Copyright (c***REMOVED*** 2023 Red Hat, Inc.

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

***REMOVED***                      // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Cluster Autoscaler", func(***REMOVED*** {
	Context("creation", func(***REMOVED*** {
		It("fails if given an out-of-range utilization threshold", func(***REMOVED*** {
			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					scale_down = {
						utilization_threshold = "1.1"
			***REMOVED***
		***REMOVED***
			`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("fails if given an invalid range", func(***REMOVED*** {
			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					resource_limits = {
						cores = {
							min = 1
							max = 0
				***REMOVED***
			***REMOVED***
		***REMOVED***
			`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("fails if given an invalid duration string", func(***REMOVED*** {
			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					max_node_provision_time = "1"
		***REMOVED***
			`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("fails to find a matching cluster object", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Cluster '123' not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
		***REMOVED***
			`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("fails if OCM backend fails to create the object", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusInternalServerError, "Internal Server Error"***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("successfully creates a cluster-autoscaler object", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					VerifyJQ(".balance_similar_node_groups", true***REMOVED***,
					VerifyJQ(".skip_nodes_with_local_storage", true***REMOVED***,
					VerifyJQ(".log_verbosity", float64(3***REMOVED******REMOVED***,
					VerifyJQ(".max_pod_grace_period", float64(1***REMOVED******REMOVED***,
					VerifyJQ(".pod_priority_threshold", float64(-10***REMOVED******REMOVED***,
					VerifyJQ(".ignore_daemonsets_utilization", false***REMOVED***,
					VerifyJQ(".max_node_provision_time", "1h"***REMOVED***,
					VerifyJQ(".balancing_ignored_labels", []interface{}{"l1", "l2"}***REMOVED***,
					VerifyJQ(".resource_limits.max_nodes_total", float64(20***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.cores.min", float64(0***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.cores.max", float64(1***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.memory.min", float64(2***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.memory.max", float64(3***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.gpus[0].type", "nvidia"***REMOVED***,
					VerifyJQ(".resource_limits.gpus[0].range.min", float64(0***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.gpus[0].range.max", float64(1***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.gpus[1].type", "intel"***REMOVED***,
					VerifyJQ(".resource_limits.gpus[1].range.min", float64(2***REMOVED******REMOVED***,
					VerifyJQ(".resource_limits.gpus[1].range.max", float64(3***REMOVED******REMOVED***,
					VerifyJQ(".scale_down.enabled", true***REMOVED***,
					VerifyJQ(".scale_down.utilization_threshold", "0.4"***REMOVED***,
					VerifyJQ(".scale_down.unneeded_time", "3h"***REMOVED***,
					VerifyJQ(".scale_down.delay_after_add", "4h"***REMOVED***,
					VerifyJQ(".scale_down.delay_after_delete", "5h"***REMOVED***,
					VerifyJQ(".scale_down.delay_after_failure", "6h"***REMOVED***,
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "ClusterAutoscaler",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

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
				***REMOVED***
						memory = {
							min = 2
							max = 3
				***REMOVED***
						gpus = [
							{
								type = "nvidia"
								range = {
									min = 0
									max = 1
						***REMOVED***
					***REMOVED***,
							{
								type = "intel"
								range = {
									min = 2
									max = 3
						***REMOVED***
					***REMOVED***
						]
			***REMOVED***
					scale_down = {
						enabled = true
						utilization_threshold = "0.4"
						unneeded_time = "3h"
						delay_after_add = "4h"
						delay_after_delete = "5h"
						delay_after_failure = "6h"
			***REMOVED***
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("importing", func(***REMOVED*** {
		It("fails if resource does not exist in OCM", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Autoscaler for cluster ID '123' is not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Import("rhcs_cluster_autoscaler.cluster_autoscaler", "123"***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("succeeds if resource exists in OCM", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"scale_down": {
								"delay_after_add": "2h"
					***REMOVED***
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"scale_down": {
								"delay_after_add": "2h"
					***REMOVED***
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Import("rhcs_cluster_autoscaler.cluster_autoscaler", "123"***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

			actualResource, ok := terraform.Resource("rhcs_cluster_autoscaler", "cluster_autoscaler"***REMOVED***.(map[string]interface{}***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED***, "Type conversion failed for the received resource state"***REMOVED***

			Expect(actualResource["attributes"]***REMOVED***.To(Equal(
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
			***REMOVED***,
		***REMOVED***,
			***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("updating", func(***REMOVED*** {
		BeforeEach(func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					VerifyJQ(".balance_similar_node_groups", true***REMOVED***,
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "ClusterAutoscaler",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("successfully applies the changes in OCM", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					VerifyJQ(".balance_similar_node_groups", true***REMOVED***,
					VerifyJQ(".skip_nodes_with_local_storage", true***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true,
							"skip_nodes_with_local_storage": true
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
					skip_nodes_with_local_storage = true
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

			actualResource, ok := terraform.Resource("rhcs_cluster_autoscaler", "cluster_autoscaler"***REMOVED***.(map[string]interface{}***REMOVED***
			Expect(ok***REMOVED***.To(BeTrue(***REMOVED***, "Type conversion failed for the received resource state"***REMOVED***

			Expect(actualResource["attributes"]***REMOVED***.To(Equal(
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
		***REMOVED***,
			***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("deletion", func(***REMOVED*** {
		BeforeEach(func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "Cluster",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123",
							"name": "cluster",
							"state": "ready"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					VerifyJQ(".balance_similar_node_groups", true***REMOVED***,
					RespondWithJSON(http.StatusCreated, `
						{
							"kind": "ClusterAutoscaler",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				resource "rhcs_cluster_autoscaler" "cluster_autoscaler" {
					cluster = "123"
					balance_similar_node_groups = true
		***REMOVED***
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("trivially succeeds if the autoscaler object does not exist in OCM", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, `
						{
							"kind": "Error",
							"id": "404",
							"href": "/api/clusters_mgmt/v1/errors/404",
							"code": "CLUSTERS-MGMT-404",
							"reason": "Autoscaler for cluster ID '123' is not found",
							"operation_id": "96ae3bc2-dd56-4640-8092-4703c81ad2c1"
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("successfully applies the deletion in OCM", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusOK, `
						{
							"kind": "ClusterAutoscaler",
							"href": "/api/clusters_mgmt/v1/clusters/123/autoscaler",
							"balance_similar_node_groups": true
				***REMOVED***
					`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/autoscaler"***REMOVED***,
					RespondWithJSON(http.StatusNoContent, "{}"***REMOVED***,
				***REMOVED***,
			***REMOVED***

			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
