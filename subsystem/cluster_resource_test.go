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

var _ = Describe("Cluster creation", func() {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	const template = `{
	  "id": "123",
	  "product": {
		"id": "osd"
	  },
	  "name": "my-cluster",
	  "cloud_provider": {
	    "id": "aws"
	  },
	  "region": {
	    "id": "us-west-1"
	  },
	  "multi_az": false,
	  "properties": {},
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
	  "nodes": {
	    "compute": 3,
        "availability_zones": ["az"],
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
	  },
	  "ccs": {
	    "enabled": false
	  },
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  "version": {
		  "id": "openshift-4.8.0"
	  },
	  "state": "ready"
	}`

	It("Creates basic cluster", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJQ(`.name`, "my-cluster"),
				VerifyJQ(`.cloud_provider.id`, "aws"),
				VerifyJQ(`.region.id`, "us-west-1"),
				RespondWithJSON(http.StatusCreated, template),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "osd"
		    cloud_provider = "aws"
		    cloud_region   = "us-west-1"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())
	})

	It("Saves API and console URLs to the state", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				RespondWithJSON(http.StatusCreated, template),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "osd"
		    cloud_provider = "aws"
		    cloud_region   = "us-west-1"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_cluster", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.api_url", "https://my-api.example.com"))
		Expect(resource).To(MatchJQ(".attributes.console_url", "https://my-console.example.com"))
	})

	It("Sets compute nodes and machine type", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJQ(`.nodes.compute`, 3.0),
				VerifyJQ(`.nodes.compute_machine_type.id`, "r5.xlarge"),
				RespondWithJSON(http.StatusCreated, template),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
		    name                 = "my-cluster"
			product		   		 = "osd"
		    cloud_provider       = "aws"
		    cloud_region         = "us-west-1"
		    compute_nodes        = 3
		    compute_machine_type = "r5.xlarge"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_cluster", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.compute_nodes", 3.0))
		Expect(resource).To(MatchJQ(".attributes.compute_machine_type", "r5.xlarge"))
	})

	It("Creates CCS cluster", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJQ(".ccs.enabled", true),
				VerifyJQ(".aws.account_id", "123"),
				VerifyJQ(".aws.access_key_id", "456"),
				VerifyJQ(".aws.secret_access_key", "789"),
				RespondWithPatchedJSON(http.StatusOK, template, `[
				  {
				    "op": "replace",
				    "path": "/ccs",
				    "value": {
				      "enabled": true
				    }
				  },
				  {
				    "op": "add",
				    "path": "/aws",
				    "value": {
				      "account_id": "123",
				      "access_key_id": "456",
				      "secret_access_key": "789"
				    }
				  }
				]`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
		    name                  = "my-cluster"
			product		   		  = "osd"
		    cloud_provider        = "aws"
		    cloud_region          = "us-west-1"
		    ccs_enabled           = true
		    aws_account_id        = "123"
		    aws_access_key_id     = "456"
		    aws_secret_access_key = "789"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_cluster", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.ccs_enabled", true))
		Expect(resource).To(MatchJQ(".attributes.aws_account_id", "123"))
		Expect(resource).To(MatchJQ(".attributes.aws_access_key_id", "456"))
		Expect(resource).To(MatchJQ(".attributes.aws_secret_access_key", "789"))
	})

	It("Sets network configuration", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJQ(".network.machine_cidr", "10.0.0.0/15"),
				VerifyJQ(".network.service_cidr", "172.30.0.0/15"),
				VerifyJQ(".network.pod_cidr", "10.128.0.0/13"),
				VerifyJQ(".network.host_prefix", 22.0),
				RespondWithPatchedJSON(http.StatusOK, template, `[
				  {
				    "op": "replace",
				    "path": "/network",
				    "value": {
				      "machine_cidr": "10.0.0.0/15",
				      "service_cidr": "172.30.0.0/15",
				      "pod_cidr": "10.128.0.0/13",
				      "host_prefix": 22
				    }
				  }
				]`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "osd"
		    cloud_provider = "aws"
		    cloud_region   = "us-west-1"
		    machine_cidr   = "10.0.0.0/15"
		    service_cidr   = "172.30.0.0/15"
		    pod_cidr       = "10.128.0.0/13"
		    host_prefix    = 22
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_cluster", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.machine_cidr", "10.0.0.0/15"))
		Expect(resource).To(MatchJQ(".attributes.service_cidr", "172.30.0.0/15"))
		Expect(resource).To(MatchJQ(".attributes.pod_cidr", "10.128.0.0/13"))
		Expect(resource).To(MatchJQ(".attributes.host_prefix", 22.0))
	})

	It("Sets version", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				VerifyJQ(".version.id", "openshift-v4.8.1"),
				RespondWithPatchedJSON(http.StatusOK, template, `[
				  {
				    "op": "replace",
				    "path": "/version",
				    "value": {
				      "id": "openshift-v4.8.1"
				    }
				  }
				]`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "osd"
		    cloud_provider = "aws"
		    cloud_region   = "us-west-1"
		    version        = "openshift-v4.8.1"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_cluster", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.version", "openshift-v4.8.1"))
	})

	It("Fails if the cluster already exists", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				RespondWithJSON(http.StatusBadRequest, `{
				  "id": "400",
				  "code": "CLUSTERS-MGMT-400",
				  "reason": "Cluster 'my-cluster' already exists"
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_cluster" "my_cluster" {
			product		   = "osd"
		    name           = "my-cluster"
		    cloud_provider = "aws"
		    cloud_region   = "us-west-1"
		  }
		`)
		Expect(terraform.Apply()).ToNot(BeZero())
	})
})
