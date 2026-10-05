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

package classic

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
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
			"ec2_metadata_http_tokens": "optional",
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
			"id": "4.10.0",
			"raw_id": "4.10.0"
		}
	}`
	Context("rhcs_cluster_rosa_classic - import", func() {
		It("leaves replicas null on import with empty config", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "ec2_metadata_http_tokens": "optional",
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
			)

			Terraform.Source(`
			  resource "rhcs_cluster_rosa_classic" "my_cluster" { }
			`)
			runOutput := Terraform.Import("rhcs_cluster_rosa_classic.my_cluster", "123")
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.10.0"))
			Expect(resource).To(MatchJQ(".attributes.replicas", nil))
		})

		It("leaves replicas null on import when compute is inflated (two pools)", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "ec2_metadata_http_tokens": "optional",
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
							"compute": 15,
							"compute_machine_type": {
								"id": "r5.xlarge"
							}
						  }
						}]`),
				),
			)

			Terraform.Source(`
			  resource "rhcs_cluster_rosa_classic" "my_cluster" { }
			`)
			runOutput := Terraform.Import("rhcs_cluster_rosa_classic.my_cluster", "123")
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.replicas", nil))
		})

		It("leaves replicas null on import when compute is high and no worker pool exists", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
						{
						  "op": "add",
						  "path": "/nodes",
						  "value": {
							"availability_zones": ["us-west-1a"],
							"compute": 15,
							"compute_machine_type": {"id": "r5.xlarge"}
						  }
						}]`),
				),
			)
			Terraform.Source(`resource "rhcs_cluster_rosa_classic" "my_cluster" {}`)
			Expect(Terraform.Import("rhcs_cluster_rosa_classic.my_cluster", "123").ExitCode).To(BeZero())
			Expect(Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")).
				To(MatchJQ(".attributes.replicas", nil))
		})

		It("leaves replicas null on import when worker pool uses autoscaling", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
						{
						  "op": "add",
						  "path": "/nodes",
						  "value": {
							"availability_zones": ["us-west-1a"],
							"compute": 15,
							"compute_machine_type": {"id": "r5.xlarge"}
						  }
						}]`),
				),
			)
			Terraform.Source(`resource "rhcs_cluster_rosa_classic" "my_cluster" {}`)
			Expect(Terraform.Import("rhcs_cluster_rosa_classic.my_cluster", "123").ExitCode).To(BeZero())
			Expect(Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")).
				To(MatchJQ(".attributes.replicas", nil))
		})

	})
})
