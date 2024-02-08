package hcp

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Machine pool delete", func() {
	clusterId := "123"

	prepareClusterRead := func(clusterId string) {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId),
				RespondWithJSONTemplate(http.StatusOK, `{
				  "id": "{{.ClusterId}}",
				  "name": "my-cluster",
				  "multi_az": true,
				  "nodes": {
					"availability_zones": [
					  "us-east-1a",
					  "us-east-1b",
					  "us-east-1c"
					]
				  },
				  "state": "ready"
				}`, "ClusterId", clusterId),
			),
		)
	}

	preparePoolRead := func(clusterId string, poolId string) {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools/"+poolId),
				RespondWithJSONTemplate(http.StatusOK, `
			{
				"id": "{{.PoolId}}",
				"kind": "NodePool",
				"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/node_pools/{{.PoolId}}",
				"replicas": 3,
				"aws_node_pool":{
					"instance_type":"r5.xlarge",
					"instance_profile": "bla"
				},
				"version": {
					"raw_id": "4.14.10"
				},
				"subnet": "subnet-123"
			}`, "PoolId", poolId, "ClusterId", clusterId),
			),
		)
	}

	createPool := func(clusterId string, poolId string) {
		prepareClusterRead(clusterId)
		prepareClusterRead(clusterId)
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools",
				),
				RespondWithJSONTemplate(http.StatusOK, `{
				  	"id": "{{.PoolId}}",
				  	"name": "{{.PoolId}}",
				  	"aws_node_pool":{
					 	"instance_type":"r5.xlarge",
					 	"instance_profile": "bla"
				  	},
				  	"replicas": 3,
				  	"availability_zone": "us-east-1a",
					"version": {
						"raw_id": "4.14.10"
					},
					"subnet": "subnet-123"
				}`, "PoolId", poolId),
			),
		)

		terraform.Source(EvaluateTemplate(`
		resource "rhcs_hcp_machine_pool" "{{.PoolId}}" {
		  	cluster      = "{{.ClusterId}}"
		  	name         = "{{.PoolId}}"
		  	aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			replicas     = 3
			subnet_id = "subnet-123"
			autoscaling = {
				enabled = false
			}
			version = "4.14.10"
		}`, "PoolId", poolId, "ClusterId", clusterId))

		// Run the apply command:
		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_hcp_machine_pool", poolId)
		Expect(resource).To(MatchJQ(".attributes.cluster", clusterId))
		Expect(resource).To(MatchJQ(".attributes.id", poolId))
		Expect(resource).To(MatchJQ(".attributes.name", poolId))
	}

	BeforeEach(func() {
		createPool(clusterId, "pool1")
	})

	It("can delete a machine pool", func() {
		// Prepare for refresh (Read) of the pools prior to changes
		preparePoolRead(clusterId, "pool1")
		// Prepare for the delete of pool1
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools/pool1"),
				RespondWithJSON(http.StatusOK, `{}`),
			),
		)

		// Re-apply w/ empty source so that pool1 is deleted
		terraform.Source("")
		Expect(terraform.Apply()).To(BeZero())
	})
	It("will return an error if delete fails and not the last pool", func() {
		// Prepare for refresh (Read) of the pools prior to changes
		preparePoolRead(clusterId, "pool1")
		// Prepare for the delete of pool1
		server.AppendHandlers(
			CombineHandlers( // Fail the delete
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools/pool1"),
				RespondWithJSON(http.StatusBadRequest, `{}`), // XXX Fix description
			),
			CombineHandlers( // List returns more than 1 pool
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools"),
				RespondWithJSONTemplate(http.StatusOK, `{
					"kind": "NodePoolList",
					"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/node_pools",
					"page": 1,
					"size": 2,
					"total": 2,
					"items": [
					  {
						"kind": "NodePool",
						"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/node_pools/worker",
						"id": "worker",
						"replicas": 2,
						"availability_zone": "us-east-1a",
						"aws_node_pool":{
						   "instance_type":"r5.xlarge",
						   "instance_profile": "bla"
						},
						"version": {
							"raw_id": "4.14.10"
						},
						"subnet": "subnet-123"
					  },
					  {
						"kind": "NodePool",
						"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/node_pools/pool1",
						"id": "pool1",
						"replicas": 2,
						"availability_zone": "us-east-1a",
						"aws_node_pool":{
						   "instance_type":"r5.xlarge",
						   "instance_profile": "bla"
						},
						"version": {
							"raw_id": "4.14.10"
						},
						"subnet": "subnet-123"
					  }
					]
				  }`),
			),
		)

		// Re-apply w/ empty source so that pool1 is (attempted) deleted
		terraform.Source("")
		Expect(terraform.Apply()).NotTo(BeZero())
	})
	It("will ignore the error if delete fails and is the last pool", func() {
		// Prepare for refresh (Read) of the pools prior to changes
		preparePoolRead(clusterId, "pool1")
		// Prepare for the delete of pool1
		server.AppendHandlers(
			CombineHandlers( // Fail the delete
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools/pool1"),
				RespondWithJSON(http.StatusBadRequest, `{}`), // XXX Fix description
			),
			CombineHandlers( // List returns only 1 pool
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/node_pools"),
				RespondWithJSONTemplate(http.StatusOK, `{
					"kind": "NodePoolList",
					"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/node_pools",
					"page": 1,
					"size": 1,
					"total": 1,
					"items": [
					  {
						"kind": "NodePool",
						"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/node_pools/pool1",
						"id": "pool1",
						"replicas": 2,
						"availability_zone": "us-east-1a",
						"aws_node_pool":{
						   "instance_type":"r5.xlarge",
						   "instance_profile": "bla"
						},
						"version": {
							"raw_id": "4.14.10"
						},
						"subnet": "subnet-123"
					  }
					]
				  }`),
			),
		)

		// Re-apply w/ empty source so that pool1 is (attempted) deleted
		terraform.Source("")
		// Last pool, we ignore the error, so this succeeds
		Expect(terraform.Apply()).To(BeZero())
	})
})
