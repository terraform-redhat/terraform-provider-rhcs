package hcp

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

const (
	workerNodePoolUri = "/api/clusters_mgmt/v1/clusters/123/node_pools/worker"
)

var _ = Describe("Day-1 machine pool (worker)", func() {
	BeforeEach(func() {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "123",
					"name": "my-cluster",
					"multi_az": false,
					"nodes": {
						"availability_zones": [
							"us-east-1a"
						]
					},
					"state": "ready"
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "123",
					"name": "my-cluster",
					"multi_az": false,
					"nodes": {
						"availability_zones": [
							"us-east-1a"
						]
					},
					"state": "ready"
				}`),
			),
		)
	})

	It("cannot be created", func() {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, workerNodePoolUri),
				RespondWithJSON(http.StatusNotFound, `
				{
					"kind": "Error",
					"id": "404",
					"href": "/api/clusters_mgmt/v1/errors/404",
					"code": "CLUSTERS-MGMT-404",
					"reason": "Node pool with id 'worker' not found.",
					"operation_id": "df359e0c-b1d3-4feb-9b58-50f7a20d0096"
				}`),
			),
		)
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "worker" {
			cluster      = "123"
			name         = "worker"
			aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
			version = "4.14.10"
			replicas     = 2
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})

	It("is automatically imported and updates applied", func() {
		// Import automatically "Create()", and update the # of replicas: 2 -> 4
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, workerNodePoolUri),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "worker",
					"replicas": 2,
					"aws_node_pool":{
						"instance_type":"r5.xlarge",
						"instance_profile": "bla"
					},
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123"
				}`),
			),
			// Get is for the read during update
			CombineHandlers(
				VerifyRequest(http.MethodGet, workerNodePoolUri),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "worker",
					"replicas": 2,
					"aws_node_pool":{
						"instance_type":"r5.xlarge",
						"instance_profile": "bla"
					},
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123"
				}`),
			),
			// Patch is for the update
			CombineHandlers(
				VerifyRequest(http.MethodPatch, workerNodePoolUri),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "worker",
					"aws_node_pool":{
						"instance_type":"r5.xlarge"
					},
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123",
					"replicas": 4
				}`),
			),
		)
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "worker" {
			cluster      = "123"
			name         = "worker"
			aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
			subnet_id = "subnet-123"
			version = "4.14.10"
			replicas     = 4
		}`)
		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_hcp_machine_pool", "worker")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.name", "worker"))
		Expect(resource).To(MatchJQ(".attributes.id", "worker"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 4.0))
	})

	It("can update labels", func() {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, workerNodePoolUri),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "worker",
					"replicas": 2,
					"aws_node_pool":{
						"instance_type":"r5.xlarge",
						"instance_profile": "bla"
					},
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123"
				}`),
			),
			// Get is for the read during update
			CombineHandlers(
				VerifyRequest(http.MethodGet, workerNodePoolUri),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "worker",
					"replicas": 2,
					"aws_node_pool":{
						"instance_type":"r5.xlarge",
						"instance_profile": "bla"
					},
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123"
				}`),
			),
			// Patch is for the update
			CombineHandlers(
				VerifyRequest(http.MethodPatch, workerNodePoolUri),
				RespondWithJSON(http.StatusOK, `
				{
					"id": "worker",
					"labels": {
						"label_key1": "label_value1"
					},
					"replicas": 2,
					"aws_node_pool":{
						"instance_type":"r5.xlarge",
						"instance_profile": "bla"
					},
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123"
				}`),
			),
		)
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "worker" {
			cluster      = "123"
			name         = "worker"
			aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
			replicas     = 2
			labels = {
				"label_key1" = "label_value1"
			}
			subnet_id = "subnet-123"
			version = "4.14.10"
		}`)
		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_hcp_machine_pool", "worker")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.name", "worker"))
		Expect(resource).To(MatchJQ(".attributes.id", "worker"))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 1))
	})
})
