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

	const templateReadyState = `{
	  "id": "123",
	  "name": "my-cluster",
	  "state": "ready",
	  "region": {
	    "id": "us-west-1"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  "version": {
		  "id": "openshift-4.8.0"
	  }
	}`

	const templateWaitingState = `{
	  "id": "123",
	  "name": "my-cluster",
	  "state": "waiting",
	  "region": {
	    "id": "us-west-1"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  "version": {
		  "id": "openshift-4.8.0"
	  }
	}`

	const templateErrorState = `{
	  "id": "123",
	  "name": "my-cluster",
	  "state": "error",
	  "region": {
	    "id": "us-west-1"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  "version": {
		  "id": "openshift-4.8.0"
	  }
	}`

	Context("Create cluster waiter resource", func() {
		It("Create cluster waiter without a timeout", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, templateReadyState),
				),
			)
			terraform.Source(`
				resource "rhcs_cluster_wait" "rosa_cluster" {
				  cluster = "123"
				}
			`)

			Expect(terraform.Apply()).To(BeZero())
		})

		It("Create cluster with a negative timeout", func() {
			terraform.Source(`
				resource "rhcs_cluster_wait" "rosa_cluster" {
				  cluster = "123"
				  timeout = -1
				}
			`)

			// it should throw an error so exit code will not be "0":
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Create cluster with a positive timeout", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, templateReadyState),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, templateReadyState),
				),
			)
			terraform.Source(`
				resource "rhcs_cluster_wait" "rosa_cluster" {
				  cluster = "123"
				  timeout = 1
				}
			`)

			Expect(terraform.Apply()).To(BeZero())
			Expect(terraform.Destroy()).To(BeZero())
		})

	})

	It("Create cluster with a positive timeout but get cluster not ready", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, templateWaitingState),
			),
		)

		terraform.Source(`
				resource "rhcs_cluster_wait" "rosa_cluster" {
				  cluster = "123"
				  timeout = 1
				}
			`)

		Expect(terraform.Apply()).To(BeZero())
		resource := terraform.Resource("rhcs_cluster_wait", "rosa_cluster")
		Expect(resource).To(MatchJQ(`.attributes.ready`, false))
	})

	It("Create cluster with a positive timeout and failed cause cluster in error state", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, templateErrorState),
			),
		)

		terraform.Source(`
				resource "rhcs_cluster_wait" "rosa_cluster" {
				  cluster = "123"
				  timeout = 1
				}
			`)

		Expect(terraform.Apply()).ToNot(BeZero())
		resource := terraform.Resource("rhcs_cluster_wait", "rosa_cluster")
		Expect(resource).To(MatchJQ(`.attributes.ready`, false))
	})
})
