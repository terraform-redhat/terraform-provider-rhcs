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
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				),
				RespondWithJSON(http.StatusNotFound, "{}"),
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
				  "replicas": 12,
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
                  "replicas": 12,
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
				  "replicas": 12,
                  "labels": {}
				}`),
			),
		)

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
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
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
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"),
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
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
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
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
				  	"max_replicas": 3,
				  	"min_replicas": 0
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
})

var _ = Describe("Machine pool w/ mAZ cluster", func() {
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
	BeforeEach(func() {
		// The first thing that the provider will do for any operation on machine pools
		// is checking that the cluster is ready, so we always need to prepare the server to
		// respond to that:
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

	It("Fails to create pool if subnet_id not in byo vpc subnets", func() {
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

var _ = Describe("Machine pool import", func() {
	It("Can import a machine pool", func() {
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
