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

var _ = Describe("Machine pool (static) validation", func() {
	It("fails if cluster ID is empty", func() {
		terraform.Source(`
		resource "rhcs_machine_pool" "my_pool" {
				cluster = ""
			}
		`)
		Expect(terraform.Apply()).ToNot(BeZero())
	})
	It("is invalid to specify both availability_zone and subnet_id", func() {
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			multi_availability_zone = true
			availability_zone = "us-east-1a"
			subnet_id = "subnet-123"
		  }
		`)
		Expect(terraform.Validate()).NotTo(BeZero())
	})

	It("is necessary to specify both min and max replicas", func() {
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
			auto_scaling = true
			min_replicas = 1
		  }
		`)
		Expect(terraform.Validate()).NotTo(BeZero())

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
			auto_scaling = true
			max_replicas = 5
		  }
		`)
		Expect(terraform.Validate()).NotTo(BeZero())
	})

	It("is invalid to specify min_replicas or max_replicas and replicas", func() {
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
			auto_scaling = true
			min_replicas = 1
			replicas     = 5
		  }
		`)
		Expect(terraform.Validate()).NotTo(BeZero())
	})
})

var _ = Describe("Machine pool creation", func() {
	prepareClusterRead := func(clusterId string) {
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
				  "state": "ready",
				  "aws": {
					"tags": {
						"cluster-tag": "cluster-value"
					}
				  }
				}`),
			),
		)
	}
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		prepareClusterRead("123")
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
			  	  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
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
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
	})

	It("Can create machine pool with compute nodes when 404 (not found)", func() {
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
			  	  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
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
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))

		// Prepare the server for update
		prepareClusterRead("123")
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				),
				RespondWithJSON(http.StatusNotFound, "{}"),
			),
		)
		prepareClusterRead("123")
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
			  	  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
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
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
	})

	It("Can create machine pool with compute nodes and update labels", func() {
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
				  "replicas": 12
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))

		// Update - change lables
		prepareClusterRead("123")
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "instance_type": "r5.xlarge"
				}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "root_volume": {
					"aws": {
					  "size": 200
					}
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
				  "replicas": 12,
				  "labels": {
				    "label_key3": "label_value3"
				  }
				}`),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "replicas": 12,
				  "labels": {
				    "label_key3": "label_value3"
				  }
				}`),
			),
		)

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			labels = {
				"label_key3" = "label_value3"
			}
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 1))

		// Update - delete lables
		prepareClusterRead("123")
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "instance_type": "r5.xlarge"
				}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
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
				  "replicas": 12,
                  "labels": {}
				}`),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "replicas": 12,
                  "labels": null
				}`),
			),
		)

		// Invalid deletion - labels map can't be empty
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
            labels       = {}
		  }
		`)
		Expect(terraform.Apply()).ToNot(BeZero())
		// Valid deletion
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 0))
	})

	It("Can create machine pool with compute nodes and update taints", func() {
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				}
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 1))

		prepareClusterRead("123")
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "instance_type": "r5.xlarge"
				}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
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
				  "replicas": 12,
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
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
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
				  ]
				}`),
			),
		)

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
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
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 2))
	})

	It("Can create machine pool with compute nodes and remove taints", func() {
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				}
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 1))

		prepareClusterRead("123")
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
				  "instance_type": "r5.xlarge"
				}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ],
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
				  "replicas": 12,
                  "taints": []
				}`),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
				}`),
			),
		)

		// invalid removal of taints
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
            taints       = []
		  }
		`)

		Expect(terraform.Apply()).ToNot(BeZero())

		// valid removal of taints
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 0))
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
				  	"max_replicas": 3,
				  	"min_replicas": 0
				  },
				  "instance_type": "r5.xlarge"
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "autoscaling": {
				    "max_replicas": 3,
				    "min_replicas": 0
				  }
				}`),
			),
		)

		// Run the apply command to create the machine pool resource:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    autoscaling_enabled = "true"
		    min_replicas = "0"
		    max_replicas = "3"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.autoscaling_enabled", true))
		Expect(resource).To(MatchJQ(".attributes.min_replicas", float64(0)))
		Expect(resource).To(MatchJQ(".attributes.max_replicas", float64(3)))

		prepareClusterRead("123")
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
				  	"max_replicas": 3,
				  	"min_replicas": 0
				  },
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "instance_type": "r5.xlarge"
				}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
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
				  	"max_replicas": 3,
				  	"min_replicas": 0
				  },
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
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
				  "replicas": 12
				}`),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
				}`),
			),
		)
		// Run the apply command to update the machine pool:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", float64(12)))
	})

	It("Can't create machine pool with compute nodes using spot instances with negative max spot price", func() {
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
						"max_price": -10
					}
				  },
				  "instance_type": "r5.xlarge",
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "aws": {
					"spot_market_options": {
						"max_price": -10
					}
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-spot-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			use_spot_instances = "true"
            max_spot_price = -10
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				},
		    ]
		  }
		`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})

	It("Can create machine pool with compute nodes and use_spot_instances false", func() {
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    use_spot_instances = "false"
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
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "aws": {
					"spot_market_options": {
						"max_price": 0.5
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-spot-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
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
					schedule_type = "NoSchedule",
				},
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 1))
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
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "aws": {
					"spot_market_options": {
					}
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-spot-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			use_spot_instances = "true"
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
				},
		    ]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-spot-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.taints | length`, 1))
		Expect(resource).To(MatchJQ(".attributes.use_spot_instances", true))
	})

	It("Can create machine pool with custom disk size", func() {
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
				  "root_volume": {
					"aws": {
					  "size": 400
					}
				  },
				  "replicas": 12
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "root_volume": {
					"aws": {
					  "size": 400
					}
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			disk_size    = 400
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.machine_type", "r5.xlarge"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 12.0))
		Expect(resource).To(MatchJQ(".attributes.disk_size", 400.0))
	})

	It("Can create pool replacing cluster aws tags", func() {
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
				  "replicas": 3,
				  "aws": {
					"kind": "AWSMachinePool",
					"tags": {
						"cluster-tag": "mp-value"
					}
				  }
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 3,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "aws": {
					"tags": {
						"cluster-tag": "mp-value"
					}
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 3
			aws_tags 	 = {"cluster-tag": "mp-value"}
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.aws_tags.[\"cluster-tag\"]", "mp-value"))
	})

	It("Can create pool w/ new aws tags", func() {
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
				  "replicas": 3,
				  "aws": {
					"kind": "AWSMachinePool",
					"tags": {
					  "foo": "bar"
					}
				  }
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 3,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "aws": {
					"tags": {
						"foo": "bar"
					}
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 3
			aws_tags 	 = {"foo": "bar"}
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.aws_tags.[\"foo\"]", "bar"))
	})

	It("Can create pool w/ new aws tags, but cannot edit", func() {
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
				  "replicas": 3,
				  "aws": {
					"kind": "AWSMachinePool",
					"tags": {
					  "foo": "bar"
					}
				  }
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 3,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "aws": {
					"tags": {
						"foo": "bar"
					}
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 3
			aws_tags 	 = {"foo": "bar"}
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.aws_tags.[\"foo\"]", "bar"))

		prepareClusterRead("123")
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 3,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "aws": {
					"tags": {
						"foo": "bar"
					}
				  }
				}`),
			),
		)

		prepareClusterRead("123")
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 3,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "aws": {
					"tags": {
						"foo": "bar"
					}
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 3
			aws_tags 	 = {"foo": "new-bar"}
		  }
		`)
		Expect(terraform.Apply()).ToNot(BeZero())
	})
})

var _ = Describe("Machine pool w/ mAZ cluster", func() {
	prepareClusterRead := func(clusterId string) {
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
		)
	}
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		prepareClusterRead("123")
	})

	It("Can create mAZ pool", func() {
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
				  "replicas": 6
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 6,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 6
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.availability_zone", ""))
		Expect(resource).To(MatchJQ(".attributes.subnet_id", ""))
	})

	It("Can create mAZ pool, setting multi_availbility_zone", func() {
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
				  "replicas": 6
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 6,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 6
			multi_availability_zone = true
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.availability_zone", ""))
	})

	It("Fails to create mAZ pool if replicas not multiple of 3", func() {
		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 2
		  }
		`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})

	It("Can create 1AZ pool", func() {
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
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1b"
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1b"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			availability_zone = "us-east-1b"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.availability_zone", "us-east-1b"))
		Expect(resource).To(MatchJQ(".attributes.multi_availability_zone", false))
	})

	It("Can create 1AZ pool w/ multi_availability_zone", func() {
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
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			multi_availability_zone = false
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.availability_zone", "us-east-1a"))
	})
})

var _ = Describe("Machine pool w/ 1AZ cluster", func() {
	prepareClusterRead := func(clusterId string) {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `{
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
	}
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is checking that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		prepareClusterRead("123")
	})

	It("Can create 1az pool", func() {
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
				  "replicas": 4
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.availability_zone", "us-east-1a"))
	})

	It("Can create 1az pool by setting multi_availability_zone", func() {
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
				  "replicas": 4
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			multi_availability_zone = false
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.availability_zone", "us-east-1a"))
	})

	It("Fails to create pool if az supplied", func() {
		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 2
			availability_zone: "us-east-1b"
	  }
		`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})
})

var _ = Describe("Machine pool w/ 1AZ byo VPC cluster", func() {
	prepareClusterRead := func(clusterId string) {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `{
					  "id": "123",
					  "name": "my-cluster",
					  "multi_az": false,
					  "nodes": {
						"availability_zones": [
						  "us-east-1a"
						]
					  },
					  "aws": {
						"subnet_ids": [
							"id1"
						]
					},
				  "state": "ready"
					}`),
			),
		)
	}
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		prepareClusterRead("123")
	})

	It("Can create pool w/ subnet_id for byo vpc", func() {
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
				  "replicas": 4,
				  "subnets": ["id1"]
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "subnets": [
					"id1"
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			subnet_id = "id1"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.subnet_id", "id1"))
	})

	It("Can create pool w/ subnet_id  and additional security group id for byo vpc", func() {
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
				  "replicas": 4,
				  "subnets": ["id1"],
				  "aws": {
					"kind": "AWSMachinePool",
					"additional_security_group_ids": [
						"id1"
					]
				  }
				}`),
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ],
				  "subnets": [
					"id1"
				  ],
				  "aws": {
						"additional_security_group_ids": [
							"id1"
				  		]
				  }
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			subnet_id = "id1"
			aws_additional_security_group_ids = ["id1"]
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.subnet_id", "id1"))
		Expect(resource).To(MatchJQ(".attributes.aws_additional_security_group_ids.[0]", "id1"))
	})

})

var _ = Describe("Machine pool import", func() {
	prepareClusterRead := func(clusterId string) {
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
				  "state": "ready",
				  "aws": {
					"tags": {
						"cluster-tag": "cluster-value"
					}
				  }
				}`),
			),
		)
	}
	It("Can import a machine pool", func() {
		prepareClusterRead("123")
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "replicas": 12,
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "instance_type": "r5.xlarge"
				}`),
			),
		)

		// Run the import command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" { }
		`)
		Expect(terraform.Import("rhcs_machine_pool.my_pool", "123,my-pool")).To(BeZero())
		resource := terraform.Resource("rhcs_machine_pool", "my_pool")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.name", "my-pool"))
		Expect(resource).To(MatchJQ(".attributes.id", "my-pool"))
	})
})

var _ = Describe("Machine pool creation for non exist cluster", func() {
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
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			subnet_id = "not-in-vpc-of-cluster"
		  }
		`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})
})

var _ = Describe("Day-1 machine pool (worker)", func() {
	prepareClusterRead := func(clusterId string) {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, `{
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
	}
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		prepareClusterRead("123")
	})

	It("cannot be created", func() {
		prepareClusterRead("123")
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				RespondWithJSON(http.StatusNotFound, `
					{
						"kind": "Error",
						"id": "404",
						"href": "/api/clusters_mgmt/v1/errors/404",
						"code": "CLUSTERS-MGMT-404",
						"reason": "Machine pool with id 'worker' not found.",
						"operation_id": "df359e0c-b1d3-4feb-9b58-50f7a20d0096"
					}`),
			),
		)
		terraform.Source(`
			  resource "rhcs_machine_pool" "worker" {
				cluster      = "123"
			    name         = "worker"
			    machine_type = "r5.xlarge"
			    replicas     = 2
			  }
			`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})

	It("is automatically imported and updates applied", func() {
		// Import automatically "Create()", and update the # of replicas: 2 -> 4
		// Prepare the server:
		prepareClusterRead("123")
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				RespondWithJSON(http.StatusOK, `
					{
						"id": "worker",
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
						"replicas": 2,
						"instance_type": "r5.xlarge"
					}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
			// Get is for the read during update
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				RespondWithJSON(http.StatusOK, `
					{
						"id": "worker",
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
						"replicas": 2,
						"instance_type": "r5.xlarge"
					}`),
			),
			// Patch is for the update
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				VerifyJSON(`{
					  "kind": "MachinePool",
					  "id": "worker",
					  "replicas": 4
					}`),
				RespondWithJSON(http.StatusOK, `
					{
					  "id": "worker",
					  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
					  "kind": "MachinePool",
					  "instance_type": "r5.xlarge",
					  "replicas": 4
					}`),
			),
		)
		terraform.Source(`
			resource "rhcs_machine_pool" "worker" {
			  cluster      = "123"
			  name         = "worker"
			  machine_type = "r5.xlarge"
			  replicas     = 4
			}
		`)
		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_machine_pool", "worker")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.name", "worker"))
		Expect(resource).To(MatchJQ(".attributes.id", "worker"))
		Expect(resource).To(MatchJQ(".attributes.replicas", 4.0))
	})

	It("can update labels", func() {
		prepareClusterRead("123")
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				RespondWithJSON(http.StatusOK, `
						{
							"id": "worker",
							"kind": "MachinePool",
							"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
							"replicas": 2,
							"instance_type": "r5.xlarge"
						}`),
			),
		)
		prepareClusterRead("123")
		server.AppendHandlers(
			// Get is for the read during update
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				RespondWithJSON(http.StatusOK, `
						{
							"id": "worker",
							"kind": "MachinePool",
							"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
							"replicas": 2,
							"instance_type": "r5.xlarge"
						}`),
			),
			// Patch is for the update
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				VerifyJSON(`{
					  "kind": "MachinePool",
						  "id": "worker",
						  "labels": {
						    "label_key1": "label_value1"
						  },
						  "replicas": 2
						}`),
				RespondWithJSON(http.StatusOK, `
						{
						  "id": "worker",
						  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
						  "kind": "MachinePool",
						  "instance_type": "r5.xlarge",
						  "labels": {
						    "label_key1": "label_value1"
						  },
						  "replicas": 2
						}`),
			),
		)
		terraform.Source(`
			resource "rhcs_machine_pool" "worker" {
				cluster      = "123"
				name         = "worker"
				machine_type = "r5.xlarge"
				replicas     = 2
				labels = {
					"label_key1" = "label_value1"
				}
			}
			`)
		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_machine_pool", "worker")
		Expect(resource).To(MatchJQ(".attributes.cluster", "123"))
		Expect(resource).To(MatchJQ(".attributes.name", "worker"))
		Expect(resource).To(MatchJQ(".attributes.id", "worker"))
		Expect(resource).To(MatchJQ(`.attributes.labels | length`, 1))
	})

	It("can't update availability_zone", func() {
		prepareClusterRead("123")
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"),
				RespondWithJSON(http.StatusOK, `
						{
							"id": "worker",
							"kind": "MachinePool",
							"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
							"replicas": 2,
							"instance_type": "r5.xlarge",
					    "availability_zones": [
                "us-east-2b"
              ]
						}`),
			),
		)
		terraform.Source(`
			resource "rhcs_machine_pool" "worker" {
				cluster           = "123"
				name              = "worker"
				machine_type      = "r5.xlarge"
			  availability_zone = "us-east-2a"
			}
			`)
		Expect(terraform.Apply()).NotTo(BeZero())
	})
})

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
				}`,
					"ClusterId", clusterId),
			),
		)
	}

	preparePoolRead := func(clusterId string, poolId string) {
		prepareClusterRead("123")
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/"+poolId),
				RespondWithJSONTemplate(http.StatusOK, `
			{
				"id": "{{.PoolId}}",
				"kind": "MachinePool",
				"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools/{{.PoolId}}",
				"replicas": 3,
				"instance_type": "r5.xlarge"
			}`,
					"PoolId", poolId,
					"ClusterId", clusterId),
			),
		)
	}

	createPool := func(clusterId string, poolId string) {
		prepareClusterRead(clusterId)
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools",
				),
				RespondWithJSONTemplate(http.StatusOK, `{
				  "id": "{{.PoolId}}",
				  "name": "{{.PoolId}}",
				  "instance_type": "r5.xlarge",
				  "replicas": 3,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
				}`,
					"PoolId", poolId),
			),
		)

		terraform.Source(EvaluateTemplate(`
		resource "rhcs_machine_pool" "{{.PoolId}}" {
		  cluster      = "{{.ClusterId}}"
		  name         = "{{.PoolId}}"
		  machine_type = "r5.xlarge"
		  replicas     = 3
		}
	  `,
			"PoolId", poolId,
			"ClusterId", clusterId))

		// Run the apply command:
		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_machine_pool", poolId)
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
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/pool1"),
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
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/pool1"),
				RespondWithJSON(http.StatusBadRequest, `{}`), // XXX Fix description
			),
			CombineHandlers( // List returns more than 1 pool
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools"),
				RespondWithJSONTemplate(http.StatusOK, `{
					"kind": "MachinePoolList",
					"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools",
					"page": 1,
					"size": 2,
					"total": 2,
					"items": [
					  {
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools/worker",
						"id": "worker",
						"replicas": 2,
						"instance_type": "m5.xlarge",
						"availability_zones": [
						  "us-east-1a"
						],
						"root_volume": {
						  "aws": {
							"size": 300
						  }
						}
					  },
					  {
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools/pool1",
						"id": "pool1",
						"replicas": 2,
						"instance_type": "m5.xlarge",
						"availability_zones": [
						  "us-east-1a"
						],
						"root_volume": {
						  "aws": {
							"size": 300
						  }
						}
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
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/pool1"),
				RespondWithJSON(http.StatusBadRequest, `{}`), // XXX Fix description
			),
			CombineHandlers( // List returns only 1 pool
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools"),
				RespondWithJSONTemplate(http.StatusOK, `{
					"kind": "MachinePoolList",
					"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools",
					"page": 1,
					"size": 1,
					"total": 1,
					"items": [
					  {
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools/pool1",
						"id": "pool1",
						"replicas": 2,
						"instance_type": "m5.xlarge",
						"availability_zones": [
						  "us-east-1a"
						],
						"root_volume": {
						  "aws": {
							"size": 300
						  }
						}
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
