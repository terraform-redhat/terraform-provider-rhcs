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

var _ = Describe("rhcs_cluster_rosa_classic - import", func() {
	const template = `{
		"id": "123",
		"name": "my-cluster",
		"domain_prefix": "my-cluster",
		"region": {
		  "id": "us-west-1"
		},
		"aws": {
			"sts": {
				"oidc_endpoint_url": "https://127.0.0.1",
				"thumbprint": "111111",
				"role_arn": "",
				"support_role_arn": "",
				"instance_iam_roles" : {
					"master_role_arn" : "",
					"worker_role_arn" : ""
				},
				"operator_role_prefix" : "test"
			}
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
		"nodes": {
			"availability_zones": [
				"us-west-1a",
				"us-west-1b",
				"us-west-1c"
			],
			"compute": 3,
			"compute_machine_type": {
				"id": "r5.xlarge"
			}
		},
		"version": {
			"id": "4.10.0"
		}
	}`
	Context("rhcs_cluster_rosa_classic - import", func() {
		It("can import a cluster", func() {
			// Prepare the server:
			server.AppendHandlers(
				// CombineHandlers(
				// 	VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				// 	RespondWithJSON(http.StatusOK, versionListPage1),
				// ),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.1",
								  "thumbprint": "111111",
								  "role_arn": "",
								  "support_role_arn": "",
								  "instance_iam_roles" : {
									"master_role_arn" : "",
									"worker_role_arn" : ""
								  },
								  "operator_role_prefix" : "test"
							  }
						  }
						},
						{
						  "op": "add",
						  "path": "/nodes",
						  "value": {
							"availability_zones": [
								"us-west-1a",
								"us-west-1b",
								"us-west-1c"
							],
							"compute": 3,
							"compute_machine_type": {
								"id": "r5.xlarge"
							}
						  }
						}]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/aws_inquiries/oidc_thumbprint"),
					RespondWithJSON(http.StatusCreated, `{
						"href": "/api/clusters_mgmt/v1/aws_inquiries/oidc_thumbprint/",
						"thumbprint": "",
						"oidc_config_id": "",
						"cluster_id": "",
						"managed": false,
						"reusable": true
					}`),
				),
			)

			// Run the apply command:
			terraform.Source(`
			  resource "rhcs_cluster_rosa_classic" "my_cluster" { }
			`)
			Expect(terraform.Import("rhcs_cluster_rosa_classic.my_cluster", "123")).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.10.0"))
		})

	})
})
