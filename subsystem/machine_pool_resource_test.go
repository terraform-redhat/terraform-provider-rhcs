/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

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

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Machine pool (static***REMOVED*** validation", func(***REMOVED*** {
	It("is invalid to specify both availability_zone and subnet_id", func(***REMOVED*** {
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
		`***REMOVED***
		Expect(terraform.Validate(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("is necessary to specify both min and max replicas", func(***REMOVED*** {
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
			auto_scaling = true
			min_replicas = 1
		  }
		`***REMOVED***
		Expect(terraform.Validate(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
			auto_scaling = true
			max_replicas = 5
		  }
		`***REMOVED***
		Expect(terraform.Validate(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("is invalid to specify min_replicas or max_replicas and replicas", func(***REMOVED*** {
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
			auto_scaling = true
			min_replicas = 1
			replicas     = 5
		  }
		`***REMOVED***
		Expect(terraform.Validate(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool creation", func(***REMOVED*** {
	BeforeEach(func(***REMOVED*** {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
			***REMOVED***,
					"state": "ready"
				  }`***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes when 404 (not found***REMOVED***", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***

		// Prepare the server for update
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				***REMOVED***,
				RespondWithJSON(http.StatusNotFound, "{}"***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes and update labels", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "replicas": 12
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***

		// Update - change lables
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
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
			***REMOVED***
				  },
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
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
			***REMOVED***
				  },
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "replicas": 12,
				  "labels": {
				    "label_key3": "label_value3"
				  }
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
				  },
				  "replicas": 12,
				  "labels": {
				    "label_key3": "label_value3"
				  }
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			labels = {
				"label_key3" = "label_value3"
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 1***REMOVED******REMOVED***

		// Update - delete lables
		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
				  },
				  "labels": {
				    "label_key1": "label_value1",
				    "label_key2": "label_value2"
				  },
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "replicas": 12,
                  "labels": {}
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
				  },
				  "replicas": 12,
                  "labels": {}
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 0***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes and update taints", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
		***REMOVED***
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.taints | length`, 1***REMOVED******REMOVED***

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
		***REMOVED***,
				{
					key = "key2",
					value = "value2",
					schedule_type = "NoExecute",
		***REMOVED***
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.taints | length`, 2***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes and remove taints", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
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
			***REMOVED***
				  },
				  "replicas": 12,
				  "taints": [
					  {
						"effect": "NoSchedule",
						"key": "key1",
						"value": "value1"
					  }
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
		***REMOVED***
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.taints | length`, 1***REMOVED******REMOVED***

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "kind": "MachinePool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
                  "replicas": 12,
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "replicas": 12,
                  "taints": []
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `
				{
				  "id": "my-pool",
				  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				  "kind": "MachinePool",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
				  },
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.taints | length`, 0***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with autoscaling enabled and update to compute nodes", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "autoscaling": {
				  	"kind": "MachinePoolAutoscaling",
				  	"max_replicas": 3,
				  	"min_replicas": 0
				  },
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 200
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.autoscaling_enabled", true***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.min_replicas", float64(0***REMOVED******REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.max_replicas", float64(3***REMOVED******REMOVED******REMOVED***

		server.AppendHandlers(
			// First get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
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
			***REMOVED***
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Second get is for the Update function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
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
			***REMOVED***
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ],
				  "instance_type": "r5.xlarge"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(
					http.MethodPatch,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "replicas": 12
		***REMOVED***`***REMOVED***,
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
			***REMOVED***
				  },
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Run the apply command to update the machine pool:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource = terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", float64(12***REMOVED******REMOVED******REMOVED***
	}***REMOVED***

	It("Can't create machine pool with compute nodes using spot instances with negative max spot price", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-spot-pool",
				  "aws": {
					"kind": "AWSMachinePool",
					"spot_market_options": {
						"kind": "AWSSpotMarketOptions",
						"max_price": -10
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "aws": {
					"spot_market_options": {
						"max_price": -10
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			use_spot_instances = "true"
            max_spot_price = -10
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes and use_spot_instances false", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
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
		***REMOVED***`***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes using spot instances with max spot price of 0.5", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-spot-pool",
				  "aws": {
					"kind": "AWSMachinePool",
					"spot_market_options": {
						"kind": "AWSSpotMarketOptions",
						"max_price": 0.5
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "aws": {
					"spot_market_options": {
						"max_price": 0.5
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			use_spot_instances = "true"
            max_spot_price = 0.5
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-spot-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-spot-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.taints | length`, 1***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.use_spot_instances", true***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.max_spot_price", float64(0.5***REMOVED******REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with compute nodes using spot instances with max spot price of on-demand price", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-spot-pool",
				  "aws": {
					"kind": "AWSMachinePool",
					"spot_market_options": {
						"kind": "AWSSpotMarketOptions"
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-spot-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 12,
				  "aws": {
					"spot_market_options": {
			***REMOVED***
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

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
	***REMOVED***
			use_spot_instances = "true"
			taints = [
				{
					key = "key1",
					value = "value1",
					schedule_type = "NoSchedule",
		***REMOVED***,
		    ]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-spot-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-spot-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 2***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.taints | length`, 1***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.use_spot_instances", true***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create machine pool with custom disk size", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "root_volume": {
					"aws": {
					  "size": 400
			***REMOVED***
				  },
				  "replicas": 12
		***REMOVED***`***REMOVED***,
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
			***REMOVED***
				  }
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 12
			disk_size    = 400
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.machine_type", "r5.xlarge"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 12.0***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.disk_size", 400.0***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool w/ mAZ cluster", func(***REMOVED*** {
	BeforeEach(func(***REMOVED*** {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
			***REMOVED***,
					"state": "ready"
				  }`***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}***REMOVED***

	It("Can create mAZ pool", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 6
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 6,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 6
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.availability_zone", ""***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.subnet_id", ""***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create mAZ pool, setting multi_availbility_zone", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 6
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 6,
				  "availability_zones": [
					"us-east-1a",
					"us-east-1b",
					"us-east-1c"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 6
			multi_availability_zone = true
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.availability_zone", ""***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails to create mAZ pool if replicas not multiple of 3", func(***REMOVED*** {
		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 2
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create 1AZ pool", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1b"
				  ]
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1b"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			availability_zone = "us-east-1b"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.availability_zone", "us-east-1b"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.multi_availability_zone", false***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create 1AZ pool w/ multi_availability_zone", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			multi_availability_zone = false
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.availability_zone", "us-east-1a"***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool w/ 1AZ cluster", func(***REMOVED*** {
	BeforeEach(func(***REMOVED*** {
		// The first thing that the provider will do for any operation on machine pools
		// is checking that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
					"id": "123",
					"name": "my-cluster",
					"multi_az": false,
					"nodes": {
					  "availability_zones": [
						"us-east-1a"
					  ]
			***REMOVED***,
					"state": "ready"
				  }`***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}***REMOVED***

	It("Can create 1az pool", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.availability_zone", "us-east-1a"***REMOVED******REMOVED***
	}***REMOVED***

	It("Can create 1az pool by setting multi_availability_zone", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "availability_zones": [
					"us-east-1a"
				  ]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			multi_availability_zone = false
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.availability_zone", "us-east-1a"***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails to create pool if az supplied", func(***REMOVED*** {
		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 2
			availability_zone: "us-east-1b"
	  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool w/ 1AZ byo VPC cluster", func(***REMOVED*** {
	BeforeEach(func(***REMOVED*** {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
			***REMOVED***,
				  "state": "ready"
			***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
			***REMOVED***,
				  "state": "ready"
			***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
						"id": "123",
						"name": "my-cluster",
						"multi_az": false,
						"nodes": {
						  "availability_zones": [
						    "us-east-1a"
						  ]
				***REMOVED***,
						"aws": {
							"subnet_ids": [
								"id1"
							]
				***REMOVED***,
						"state": "ready"
					  }`***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}***REMOVED***

	It("Can create pool w/ subnet_id for byo vpc", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/123/machine_pools",
				***REMOVED***,
				VerifyJSON(`{
				  "kind": "MachinePool",
				  "id": "my-pool",
				  "instance_type": "r5.xlarge",
				  "replicas": 4,
				  "subnets": ["id1"]
		***REMOVED***`***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			subnet_id = "id1"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.subnet_id", "id1"***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails to create pool if subnet_id not in byo vpc subnets", func(***REMOVED*** {
		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			subnet_id = "not-in-vpc-of-cluster"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool import", func(***REMOVED*** {
	It("Can import a machine pool", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/my-pool"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the import command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" { }
		`***REMOVED***
		Expect(terraform.Import("rhcs_machine_pool.my_pool", "123,my-pool"***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_machine_pool", "my_pool"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-pool"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "my-pool"***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool creation for non exist cluster", func(***REMOVED*** {
	It("Fail to create machine pool if cluster is not exist", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusNotFound, `{}`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_machine_pool" "my_pool" {
		    cluster      = "123"
		    name         = "my-pool"
		    machine_type = "r5.xlarge"
		    replicas     = 4
			subnet_id = "not-in-vpc-of-cluster"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Day-1 machine pool (worker***REMOVED***", func(***REMOVED*** {
	BeforeEach(func(***REMOVED*** {
		// The first thing that the provider will do for any operation on machine pools
		// is check that the cluster is ready, so we always need to prepare the server to
		// respond to that:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
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
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}***REMOVED***

	It("cannot be created", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				RespondWithJSON(http.StatusNotFound, `
					{
						"kind": "Error",
						"id": "404",
						"href": "/api/clusters_mgmt/v1/errors/404",
						"code": "CLUSTERS-MGMT-404",
						"reason": "Machine pool with id 'worker' not found.",
						"operation_id": "df359e0c-b1d3-4feb-9b58-50f7a20d0096"
			***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		terraform.Source(`
			  resource "rhcs_machine_pool" "worker" {
				cluster      = "123"
			    name         = "worker"
			    machine_type = "r5.xlarge"
			    replicas     = 2
			  }
			`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("is automatically imported and updates applied", func(***REMOVED*** {
		// Import automatically "Create(***REMOVED***", and update the # of replicas: 2 -> 4
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
					{
						"id": "worker",
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
						"replicas": 2,
						"instance_type": "r5.xlarge"
			***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Get is for the read during update
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
					{
						"id": "worker",
						"kind": "MachinePool",
						"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
						"replicas": 2,
						"instance_type": "r5.xlarge"
			***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Patch is for the update
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				VerifyJSON(`{
					  "kind": "MachinePool",
					  "id": "worker",
					  "replicas": 4
			***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `
					{
					  "id": "worker",
					  "href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
					  "kind": "MachinePool",
					  "instance_type": "r5.xlarge",
					  "replicas": 4
			***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		terraform.Source(`
			resource "rhcs_machine_pool" "worker" {
			  cluster      = "123"
			  name         = "worker"
			  machine_type = "r5.xlarge"
			  replicas     = 4
	***REMOVED***
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_machine_pool", "worker"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "worker"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "worker"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.replicas", 4.0***REMOVED******REMOVED***
	}***REMOVED***

	It("can update labels", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// Get is for the Read function
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
						{
							"id": "worker",
							"kind": "MachinePool",
							"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
							"replicas": 2,
							"instance_type": "r5.xlarge"
				***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Get is for the read during update
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				RespondWithJSON(http.StatusOK, `
						{
							"id": "worker",
							"kind": "MachinePool",
							"href": "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker",
							"replicas": 2,
							"instance_type": "r5.xlarge"
				***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Patch is for the update
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/machine_pools/worker"***REMOVED***,
				VerifyJSON(`{
					  "kind": "MachinePool",
						  "id": "worker",
						  "labels": {
						    "label_key1": "label_value1"
						  },
						  "replicas": 2
				***REMOVED***`***REMOVED***,
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
				***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		terraform.Source(`
			resource "rhcs_machine_pool" "worker" {
				cluster      = "123"
				name         = "worker"
				machine_type = "r5.xlarge"
				replicas     = 2
				labels = {
					"label_key1" = "label_value1"
		***REMOVED***
	***REMOVED***
			`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_machine_pool", "worker"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "worker"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "worker"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.labels | length`, 1***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Machine pool delete", func(***REMOVED*** {
	clusterId := "123"

	prepareClusterRead := func(clusterId string***REMOVED*** {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId***REMOVED***,
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
		***REMOVED***`,
					"ClusterId", clusterId***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}

	preparePoolRead := func(clusterId string, poolId string***REMOVED*** {
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/"+poolId***REMOVED***,
				RespondWithJSONTemplate(http.StatusOK, `
			{
				"id": "{{.PoolId}}",
				"kind": "MachinePool",
				"href": "/api/clusters_mgmt/v1/clusters/{{.ClusterId}}/machine_pools/{{.PoolId}}",
				"replicas": 3,
				"instance_type": "r5.xlarge"
	***REMOVED***`,
					"PoolId", poolId,
					"ClusterId", clusterId***REMOVED***,
			***REMOVED***,
		***REMOVED***
	}

	createPool := func(clusterId string, poolId string***REMOVED*** {
		prepareClusterRead(clusterId***REMOVED***
		prepareClusterRead(clusterId***REMOVED***
		prepareClusterRead(clusterId***REMOVED***
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(
					http.MethodPost,
					"/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools",
				***REMOVED***,
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
		***REMOVED***`,
					"PoolId", poolId***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(EvaluateTemplate(`
		resource "rhcs_machine_pool" "{{.PoolId}}" {
		  cluster      = "{{.ClusterId}}"
		  name         = "{{.PoolId}}"
		  machine_type = "r5.xlarge"
		  replicas     = 3
***REMOVED***
	  `,
			"PoolId", poolId,
			"ClusterId", clusterId***REMOVED******REMOVED***

		// Run the apply command:
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_machine_pool", poolId***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.cluster", clusterId***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", poolId***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", poolId***REMOVED******REMOVED***
	}

	BeforeEach(func(***REMOVED*** {
		createPool(clusterId, "pool1"***REMOVED***
	}***REMOVED***

	It("can delete a machine pool", func(***REMOVED*** {
		// Prepare for refresh (Read***REMOVED*** of the pools prior to changes
		preparePoolRead(clusterId, "pool1"***REMOVED***
		// Prepare for the delete of pool1
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/pool1"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{}`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Re-apply w/ empty source so that pool1 is deleted
		terraform.Source(""***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
	It("will return an error if delete fails and not the last pool", func(***REMOVED*** {
		// Prepare for refresh (Read***REMOVED*** of the pools prior to changes
		preparePoolRead(clusterId, "pool1"***REMOVED***
		// Prepare for the delete of pool1
		server.AppendHandlers(
			CombineHandlers( // Fail the delete
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/pool1"***REMOVED***,
				RespondWithJSON(http.StatusBadRequest, `{}`***REMOVED***, // XXX Fix description
			***REMOVED***,
			CombineHandlers( // List returns more than 1 pool
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools"***REMOVED***,
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
				***REMOVED***
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
				***REMOVED***
					  }
					]
				  }`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Re-apply w/ empty source so that pool1 is (attempted***REMOVED*** deleted
		terraform.Source(""***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
	It("will ignore the error if delete fails and is the last pool", func(***REMOVED*** {
		// Prepare for refresh (Read***REMOVED*** of the pools prior to changes
		preparePoolRead(clusterId, "pool1"***REMOVED***
		// Prepare for the delete of pool1
		server.AppendHandlers(
			CombineHandlers( // Fail the delete
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools/pool1"***REMOVED***,
				RespondWithJSON(http.StatusBadRequest, `{}`***REMOVED***, // XXX Fix description
			***REMOVED***,
			CombineHandlers( // List returns only 1 pool
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/"+clusterId+"/machine_pools"***REMOVED***,
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
				***REMOVED***
					  }
					]
				  }`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Re-apply w/ empty source so that pool1 is (attempted***REMOVED*** deleted
		terraform.Source(""***REMOVED***
		// Last pool, we ignore the error, so this succeeds
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
