/*
Copyright (c) 2021 Red Hat, Inc.

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

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Machine pool creation", func() {
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `{
				  "id": "123",
				  "name": "my-cluster",
				  "state": "ready"
				}`),
			),
		)
	})

	It("Can create machine pool with compute nodes", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				),
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "replicas": 10,
				  "taints": [
					  {
						"effect": "effect1",
						"key": "key1",
						"value": "value1"
					  }
				 ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 10,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 10
			labels = {
				"label_key1" = "label_value1", 
				"label_key2" = "label_value2"
			}
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "effect1",
				},
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 10.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
	})

	It("Can create machine pool with autoscaling enabled and update to compute nodes", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				),
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "autoscaling": {
				  	"kind": "MachinePoolAutoscaling",
				  	"max_replicas": 2,
				  	"min_replicas": 0
				  },
				  "instance_type": "r5.xlarge"
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "autoscaling": {
				    "max_replicas": 2,
				    "min_replicas": 0	  
				  }
				}`),
			),
		)

		// Run the apply command to create the machine pool resource:
		terraform.Source(`
		  resource "ocm_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    autoscaling_enabled = "true"
		    min_replicas = "0"
		    max_replicas = "2"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.autoscaling_enabled", true))
		Expect(resource).To(MatchJQ(".attributes.min_replicas", float64(0)))
		Expect(resource).To(MatchJQ(".attributes.max_replicas", float64(2)))

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "autoscaling": {
				  	"kind": "MachinePoolAutoscaling",
				  	"max_replicas": 2,
				  	"min_replicas": 0
				  },
				  "instance_type": "r5.xlarge"
				}`),
			),
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "autoscaling": {
				  	"kind": "MachinePoolAutoscaling",
				  	"max_replicas": 2,
				  	"min_replicas": 0
				  },
				  "instance_type": "r5.xlarge"
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				),
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "replicas": 10
				}`),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "replicas": 10
				}`),
			),
		)
		// Run the apply command to update the machine pool:
		terraform.Source(`
		  resource "ocm_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 10
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("ocm_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", float64(10)))
	})

	It("Can create machine pool with compute nodes using spot instances with max spot price of 0.5", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				),
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-spot-pool",
				  "aws": {
					"kind": "AWSMachinePool",
					"spot_market_options": {
						"kind": "AWSSpotMarketOptions",
						"max_price": 0.5
					} 
				  },
				  "instance_type": "r5.xlarge",
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "replicas": 10,
				  "taints": [
					  {
						"effect": "effect1",
						"key": "key1",
						"value": "value1"
					  }
				 ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 10,
				  "aws": {    
					"spot_market_options": {      
						"max_price": 0.5    
					}  
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-spot-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 10
			labels = {
				"label_key1" = "label_value1", 
				"label_key2" = "label_value2"
			}
			use_spot_instances = "true"
            max_spot_price = 0.5
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "effect1",
				},
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 10.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
		Expect(resource).To(MatchJQ(".attributes.use_spot_instances", true))
		Expect(resource).To(MatchJQ(".attributes.max_spot_price", float64(0.5)))
	})

	It("Can create machine pool with compute nodes using spot instances with max spot price of on-demand price", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				),
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-spot-pool",
				  "aws": {
					"kind": "AWSMachinePool",
					"spot_market_options": {
						"kind": "AWSSpotMarketOptions"
					} 
				  },
				  "instance_type": "r5.xlarge",
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "replicas": 10,
				  "taints": [
					  {
						"effect": "effect1",
						"key": "key1",
						"value": "value1"
					  }
				 ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 10,
				  "aws": {    
					"spot_market_options": {      
					}  
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-spot-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 10
			labels = {
				"label_key1" = "label_value1", 
				"label_key2" = "label_value2"
			}
			use_spot_instances = "true"
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "effect1",
				},
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 10.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
		Expect(resource).To(MatchJQ(".attributes.use_spot_instances", true))
	})
})
