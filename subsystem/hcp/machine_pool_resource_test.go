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

package hcp

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Hcp Machine pool", func() {
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
				  "multi_az": true,
				  "nodes": {
					"availability_zones": [
					  "us-east-1a",
					  "us-east-1b",
					  "us-east-1c"
					]
				  },
				  "state": "ready"
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `{
				  "id": "123",
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
				}`),
			),
		)
	})

	Context("static validation", func() {
		It("is invalid to specify both availability_zone and subnet_id", func() {
			terraform.Source(`
			resource "rhcs_hcp_machine_pool" "my_pool" {
				cluster      = "123"
				name         = "my-pool"
				aws_node_pool = {
					instance_type = "r5.xlarge"
				}
				replicas     = 12
				subnet_id = "subnet-123"
			}`)
			Expect(terraform.Validate()).NotTo(BeZero())
		})

		It("is necessary to specify both min and max replicas", func() {
			terraform.Source(`
			resource "rhcs_hcp_machine_pool" "my_pool" {
				cluster      = "123"
				name         = "my-pool"
				aws_node_pool = {
					instance_type = "r5.xlarge",
				}
				autoscaling = {
					enabled = true,
					min_replicas = 1
				}
				subnet_id = "subnet-123"
			}`)
			Expect(terraform.Validate()).NotTo(BeZero())

			terraform.Source(`
			resource "rhcs_hcp_machine_pool" "my_pool" {
				cluster      = "123"
				name         = "my-pool"
				aws_node_pool = {
					instance_type = "r5.xlarge",
				}
				autoscaling = {
					enabled = true,
					max_replicas = 5
				}
				subnet_id = "subnet-123"
			}`)
			Expect(terraform.Validate()).NotTo(BeZero())
		})

		It("is invalid to specify min_replicas and replicas", func() {
			terraform.Source(`
			resource "rhcs_hcp_machine_pool" "my_pool" {
				cluster      = "123"
				name         = "my-pool"
				aws_node_pool = {
					instance_type = "r5.xlarge",
				}
				autoscaling = {
					enabled = true,
					min_replicas = 1
				}
				replicas     = 5
				subnet_id = "subnet-123"
			}`)
			Expect(terraform.Validate()).NotTo(BeZero())
		})

		It("is invalid to specify max_replicas and replicas", func() {
			terraform.Source(`
			resource "rhcs_hcp_machine_pool" "my_pool" {
				cluster = "123"
				name = "my-pool"
				aws_node_pool = {
					instance_type = "r5.xlarge",
				}
				autoscaling = {
					enabled = true,
					max_replicas = 1
				}
				replicas = 5
				subnet_id = "subnet-123"
			}`)
			Expect(terraform.Validate()).NotTo(BeZero())
		})
	})

	It("Can create machine pool with compute nodes", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusCreated, `{
					"id":"my-pool",
					"aws_node_pool":{
					   "instance_type":"r5.xlarge",
					   "instance_profile": "bla"
					},
					"auto_repair": true,
					"replicas":12,
					"labels":{
					   "label_key1":"label_value1",
					   "label_key2":"label_value2"
					},
					"subnet":"id-1",
					"availability_zone":"us-east-1a",
					"taints":[
					   {
						  "effect":"NoSchedule",
						  "key":"key1",
						  "value":"value1"
					   }
					],
					"version": {
						"raw_id": "4.14.10"
					}
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
			aws_node_pool = {
				instance_type = "r5.xlarge",
			}
			autoscaling = {
				enabled = false,
			}
			subnet_id = "id-1"
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				},
		    ]
			version = "4.14.10"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
	})

	It("Can create machine pool with compute nodes when 404 (not found)", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusCreated, `{
				  	"id": "my-pool",
				  	"aws_node_pool": {
					  	"instance_type": "r5.xlarge",
					  	"instance_profile": "bla"
				  	},
				  	"auto_repair": true,
				  	"replicas": 12,
				  	"labels": {
					    "label_key1": "label_value1",
				    	"label_key2": "label_value2"
				  	},
				  	"subnet": "id-1",
				  	"availability_zone": "us-east-1a",
			  	  	"taints": [
					  	{
							"effect": "NoSchedule",
							"key": "key1",
							"value": "value1"
					  	}
				  	],
				  	"version": {
					  	"raw_id": "4.14.10"
				  	}
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
			subnet_id = "id-1"
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				},
		    ]
			version = "4.14.10"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))

		// Prepare the server for update
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				),
				RespondWithJSON(http.StatusNotFound, "{}"),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "123",
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
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "123",
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
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusCreated, `
				{
				  	"id": "my-pool",
				  	"aws_node_pool": {
						"instance_type": "r5.xlarge",
					  	"instance_profile": "bla"
				  	},
				  	"auto_repair": true,
				  	"replicas": 12,
				  	"labels": {
					    "label_key1": "label_value1",
				    	"label_key2": "label_value2"
				  	},
				  	"subnet": "id-1",
				  	"availability_zone": "us-east-1a",
			  	  	"taints": [
					  	{
							"effect": "NoSchedule",
							"key": "key1",
							"value": "value1"
					  	}
				  	],
					"version": {
						"raw_id": "4.14.10"
					}
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
			subnet_id = "id-1"
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				},
		    ]
			version = "4.14.10"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
	})

	It("Can create machine pool with compute nodes and update labels", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "subnet": "subnet-123",
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false,
			}
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))

		// Update - change lables
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "kind": "MachinePool",
				  "replicas": 12,
				  "labels": {
				    "label_key3": "label_value3"
				  },
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
			labels = {
				"label_key3" = "label_value3"
			}
			autoscaling = {
				enabled = false,
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 1))

		// Update - delete lables
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "kind": "MachinePool",
				  "replicas": 12,
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		// Invalid deletion - labels map can't be empty
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
	        labels       = {}
			autoscaling = {
				enabled = false,
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).ToNot(BeZero())
		// Valid deletion
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
			autoscaling = {
				enabled = false,
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 0))
	})

	It("Can create machine pool with compute nodes and update taints", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false,
			}
		    replicas     = 12
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				}
		    ]
			version = "4.14.10"
			subnet_id = "subnet-123"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 1))

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "kind": "MachinePool",
				  "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  },
					  {
						"effect": "NoExecute",
						"key": "key2",
						"value": "value2"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				},
				{
					key = "key2",
					value = "value2",
					schedule_type = "NoExecute",
				}
		    ]
			autoscaling = {
				enabled = false,
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 2))
	})

	It("Can create machine pool with compute nodes and remove taints", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "availability_zone": "us-east-1a",
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
		    replicas     = 12
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				}
		    ]
			version = "4.14.10"
			subnet_id = "subnet-123"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 1))

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
	              "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "kind": "MachinePool",
				  "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		// invalid removal of taints
		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
	        taints       = []
			autoscaling = {
				enabled = false
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		  }
		`)

		Expect(terraform.Apply()).ToNot(BeZero())

		// valid removal of taints
		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
			autoscaling = {
				enabled = false
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 0))
	})

	It("Can create machine pool with autoscaling enabled and update to compute nodes", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/node_pools",
				),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "availability_zone": "us-east-1a",
				  "autoscaling": {
				    "max_replicas": 3,
				    "min_replicas": 0
				  },
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)

		// Run the apply command to create the machine pool resource:
		terraform.Source(`
		resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
			aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = true
				min_replicas = 0
				max_replicas = 3
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.autoscaling.enabled", true))
		Expect(resource).To(MatchJQ(".attributes.autoscaling.min_replicas", float64(0)))
		Expect(resource).To(MatchJQ(".attributes.autoscaling.max_replicas", float64(3)))

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "autoscaling": {
				  	"max_replicas": 3,
				  	"min_replicas": 0
				  },
				  "availability_zone": "us-east-1a",
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "autoscaling": {
				  	"max_replicas": 3,
				  	"min_replicas": 0
				  },
				  "availability_zone": "us-east-1a",
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
				  "kind": "MachinePool",
				  "replicas": 12,
				  "availability_zone": "us-east-1a",
				  "aws_node_pool": {
					"instance_type": "r5.xlarge",
					"instance_profile": "bla"
				  },
				  "auto_repair": true,
				  "version": {
					  "raw_id": "4.14.10"
				  },
				  "subnet": "subnet-123"
				}`),
			),
		)
		// Run the apply command to update the machine pool:
		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
		    replicas     = 12
			autoscaling = {
				enabled = false
			}
			version = "4.14.10"
			subnet_id = "subnet-123"
		}`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.aws_node_pool.instance_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", float64(12)))
	})
})

var _ = Describe("Import", func() {
	It("Can import a machine pool", func() {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
					{
					  "id": "my-pool",
					  "kind": "MachinePool",
					  "href": "/api/clusters_mgmt/v1/clusters/123/node_pools/my-pool",
					  "replicas": 12,
					  "labels": {
						"label_key1": "label_value1",
						"label_key2": "label_value2"
					  },
					  "aws_node_pool": {
						"instance_type": "r5.xlarge",
						"instance_profile": "bla"
					  },
					  "auto_repair": true,
					  "version": {
						  "raw_id": "4.14.10"
					  }
					}`),
			),
		)

		// Run the import command:
		terraform.Source(`resource "rhcs_hcp_machine_pool" "my_pool" {}`)
		Expect(terraform.Import("rhcs_hcp_machine_pool.my_pool", "123,my-pool")).To(BeZero())
		resource := terraform.Resource("rhcs_hcp_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
	})
})

var _ = Context("Machine pool creation for non exist cluster", func() {
	It("Fail to create machine pool if cluster is not exist", func() {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusNotFound, `{}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_hcp_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    aws_node_pool = {
				instance_type = "r5.xlarge"
			}
			autoscaling = {
				enabled = false
			}
		    replicas     = 4
			subnet_id = "not-in-vpc-of-cluster"
			version = "4.14.10"
		  }
		`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})
})
