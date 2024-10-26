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
	"fmt"
	"net/http"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

const versionListPage1 = `{
	"kind": "VersionList",
	"page": 1,
	"size": 2,
	"total": 2,
	"items": [{
			"kind": "Version",
			"id": "openshift-v4.10.1",
			"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.1",
			"raw_id": "4.10.1"
		},
		{
			"kind": "Version",
			"id": "openshift-4.14.0",
			"href": "/api/clusters_mgmt/v1/versions/openshift-4.14.0",
			"raw_id": "4.14.0"
		},
		{
			"kind": "Version",
			"id": "openshift-v4.10.1",
			"href": "/api/clusters_mgmt/v1/versions/openshift-v4.11.1",
			"raw_id": "4.11.1"
		}
	]
}`

var _ = Describe("rhcs_cluster_rosa_classic - create", func() {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	template := `{
	  "id": "123",
	  "external_id": "123",
	  "infra_id": "my-cluster-123",
	  "name": "my-cluster",
	  "domain_prefix": "mydomainprefix",
	  "state": "ready",
	  "region": {
	    "id": "us-west-1"
	  },
	  "aws": {
	    "ec2_metadata_http_tokens": "optional"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
      "properties": {
		 "rosa_creator_arn:": "arn:aws:iam::123456789012:user/dummy",
         "rosa_tf_version": "` + build.Version + `",
         "rosa_tf_commit": "` + build.Commit + `"
      },
	  "nodes": {
	    "compute": 3,
        "availability_zones": ["us-west-1a"],
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
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
      "dns" : {
          "base_domain": "mycluster-api.example.com"
      }
	}`

	templateWithTrustBundle := `{
	  "id": "123",
	  "external_id": "123",
	  "infra_id": "my-cluster-123",
	  "name": "my-cluster",
	  "domain_prefix": "mydomainprefix",
	  "additional_trust_bundle" : "REDUCTED",
	  "state": "ready",
	  "region": {
	    "id": "us-west-1"
	  },
	  "aws": {
	    "ec2_metadata_http_tokens": "optional"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
      "properties": {
         "rosa_tf_version": "` + build.Version + `",
         "rosa_tf_commit": "` + build.Commit + `"
      },
	  "nodes": {
	    "compute": 3,
        "availability_zones": ["us-west-1a"],
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
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
      "dns" : {
          "base_domain": "mycluster-api.example.com"
      }
	}`

	const templateReadyState = `{
	  "id": "123",
	  "name": "my-cluster",
	  "domain_prefix": "my-cluster",
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
	  },
      "nodes": {
	    "compute": 3,
        "availability_zones": ["us-west-1a"],
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
	  },
      "dns" : {
          "base_domain": "mycluster-api.example.com"
      }
	}`

	Context("rhcs_cluster_rosa_classic - create", func() {
		It("invalid az for region", func() {
			Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
              availability_zones = ["us-east-1a"]
			  aws_account_id = "123456789012"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  version = "4.11.1"
			}`)
			Expect(Terraform.Apply()).NotTo(BeZero())
		})

		It("version with unsupported prefix error", func() {
			Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123456789012"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  version = "openshift-v4.11.1"
			}`)
			Expect(Terraform.Apply()).NotTo(BeZero())
		})

		Context("Test channel groups", func() {
			It("doesn't append the channel group when on the default channel", func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage1),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						VerifyJQ(`.version.id`, "openshift-v4.11.1"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
							"op": "replace",
							"path": "/version/id",
							"value": "openshift-v4.11.1"
						}
						]`),
					),
				)
				Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123456789012"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  version = "4.11.1"
			}
		  `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})
			It("appends the channel group when on a non-default channel", func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithPatchedJSON(http.StatusOK, versionListPage1, `[
						{
							"op": "add",
							"path": "/items/-",
							"value": {
								"kind": "Version",
								"id": "openshift-v4.50.0-fast",
								"href": "/api/clusters_mgmt/v1/versions/openshift-v4.50.0-fast",
								"raw_id": "4.50.0",
								"channel_group": "fast"
							}
						}
					]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						VerifyJQ(`.version.id`, "openshift-v4.50.0-fast"),
						VerifyJQ(`.version.channel_group`, "fast"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
							"op": "replace",
							"path": "/version/id",
							"value": "openshift-v4.50.0-fast"
						},
						{
							"op": "add",
							"path": "/version/channel_group",
							"value": "fast"
						}
						]`),
					),
				)
				Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123456789012"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  channel_group = "fast"
			  version = "4.50.0"
			}
		  `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})
			It("returns an error when the version is not found in the channel group", func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithPatchedJSON(http.StatusOK, versionListPage1, `[
						{
							"op": "add",
							"path": "/items/-",
							"value": {
								"kind": "Version",
								"id": "openshift-v4.50.0-fast",
								"href": "/api/clusters_mgmt/v1/versions/openshift-v4.50.0-fast",
								"raw_id": "4.50.0",
								"channel_group": "fast"
							}
						}
					]`),
					),
				)
				Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123456789012"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  channel_group = "fast"
			  version = "4.99.99"
			}
		  `)
				Expect(Terraform.Apply()).NotTo(BeZero())
			})
		})

		Context("Test wait attribute", func() {
			It("Create cluster and wait till it will be in error state", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage1),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
					),
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
                      "path": "/state",
					  "value": "error"
					}]`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			wait_for_create_complete = true
		  }
		`)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("Cluster '123' is in state 'error' and will not become ready")
				resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.state", "error"))
			})

			It("Create cluster and wait till it will be in ready state", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage1),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
					),
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
                      "path": "/state",
					  "value": "ready"
					}]`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			wait_for_create_complete = true
		  }
		`)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.state", "ready"))
			})
		})
		It("Creates basic cluster", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.domain_prefix`, "mydomainprefix"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
			domain_prefix  = "mydomainprefix"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
			Expect(resource).To(MatchJQ(".attributes.infra_id", "my-cluster-123"))
		})

		It("Creates basic cluster returned empty az list", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						  "compute": 3,
	            "compute_machine_type": {
	              "id": "r5.xlarge"
	            }
					  }
					}
					]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
		})

		It("Creates basic cluster with admin user - default username", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.htpasswd.users.items[0].username`, "cluster-admin"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            create_admin_user = true
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.admin_credentials.username", "cluster-admin"))
			Expect(strings.Contains(fmt.Sprintf("%v", resource), "password")).Should(BeTrue())
		})

		It("Creates basic cluster with admin user - customized username/password", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.htpasswd.users.items[0].username`, "test-admin"),
					VerifyJQ(`.htpasswd.users.items[0].hashed_password`, "hash(1234AbB2341234)"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            admin_credentials = {
                username = "test-admin"
                password = "1234AbB2341234"
            }
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.admin_credentials.username", "test-admin"))
			Expect(resource).To(MatchJQ(".attributes.admin_credentials.password", "1234AbB2341234"))
		})

		It("Creates basic cluster with empty admincredentials and update the clustrer w/o updates on it", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.htpasswd.users.items[0].username`, "cluster-admin"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            admin_credentials = {}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.admin_credentials.username", "cluster-admin"))
			Expect(strings.Contains(fmt.Sprintf("%v", resource), "password")).Should(BeTrue())

			// Prepare server for update
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
					"path": "/properties",
					"value": {
						"rosa_tf_commit":"`+build.Commit+`",
						"rosa_tf_version":"`+build.Version+`"
					}
					}]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
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
					"path": "/properties",
					"value": {
						"rosa_tf_commit":"`+build.Commit+`",
						"rosa_tf_version":"`+build.Version+`"
					}
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123456789012"
				admin_credentials = {}
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
			  }
		`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Should fail to update create_admin_user if no admin user created in cluster creation", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(strings.Contains(fmt.Sprintf("%v", resource), "username")).Should(BeFalse())
			Expect(strings.Contains(fmt.Sprintf("%v", resource), "password")).Should(BeFalse())

			// Prepare server for update
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`"
                      }
                    }]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123456789012"
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
				# >>>> inject create_admin_user 
				create_admin_user = true
			  }
		`)
			Expect(Terraform.Apply()).NotTo(BeZero())
		})

		It("Creates basic cluster - and reconcile on a 404", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
			Expect(resource).To(MatchJQ(".attributes.id", "123")) // cluster has id 123

			// Prepare the server for reconcile
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
                    {
                      "op": "replace",
                      "path": "/id",
                      "value": "1234"
                    },
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource = Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
			Expect(resource).To(MatchJQ(".attributes.id", "1234")) // reconciled cluster has id of 1234
		})

		It("Creates basic cluster with custom worker disk size", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					VerifyJQ(`.nodes.compute_root_volume.aws.size`, 400.0),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
							"compute": 3,
							"availability_zones": ["az"],
							"compute_machine_type": {
								"id": "r5.xlarge"
							},
							"compute_root_volume": {
								"aws": {
									"size": 400
								}
							}
						  }
						}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
			  resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123456789012"
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
				worker_disk_size = 400
				version = "4.14.0"
			  }
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
			Expect(resource).To(MatchJQ(".attributes.worker_disk_size", 400.0))
		})

		It("Fails to create cluster with invalid custom disk size", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
			)

			// Run the apply command:
			Terraform.Source(`
			  resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123456789012"
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
				worker_disk_size = 20
				version = "4.14.0"
			  }
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Invalid root disk size")
		})

		It("Creates basic cluster with properties", func() {
			prop_key := "my_prop_key"
			prop_val := "my_prop_val"
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					VerifyJQ(`.properties.`+prop_key, prop_val),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			Terraform.Source(fmt.Sprintf(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            properties = { 
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy",
				%s = "%s"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`, prop_key, prop_val))
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Creates basic cluster with properties and update them", func() {
			prop_key := "my_prop_key"
			prop_val := "my_prop_val"
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					VerifyJQ(`.properties.`+prop_key, prop_val),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			Terraform.Source(fmt.Sprintf(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            properties = {
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy",
				%s = "%s"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`, prop_key, prop_val))
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+prop_key, prop_val))

			// Prepare server for update
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					VerifyJQ(`.properties.`+prop_key, prop_val+"_1"),
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`_1"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			Terraform.Source(fmt.Sprintf(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            properties = {
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy",
				%s = "%s"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`, prop_key, prop_val+"_1"))
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource = Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val+"_1"))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+prop_key, prop_val+"_1"))
		})

		It("Creates basic cluster with properties and delete them", func() {
			prop_key := "my_prop_key"
			prop_val := "my_prop_val"
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					VerifyJQ(`.properties.`+prop_key, prop_val),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			Terraform.Source(fmt.Sprintf(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            properties = {
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy",
				%s = "%s"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`, prop_key, prop_val))
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+prop_key, prop_val))
			Expect(resource).To(MatchJQ(`.attributes.properties| keys | length`, 2))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties| keys | length`, 4))

			// Prepare server for update
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					VerifyJQ(`.properties.`+prop_key, nil),
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
                      "path": "/properties",
                      "value": {
						"rosa_creator_arn": "arn:aws:iam::123456789012:user/dummy",
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            properties = {
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource = Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.properties | keys | length`, 1))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties | keys | length`, 3))
		})

		It("Should fail cluster creation when trying to override reserved properties", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
				),
			)
			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			properties = {
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy",
   				rosa_tf_version = "bob"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Can not override reserved properties keys. rosa_tf_version is a reserved property key")
		})

		It("Should fail cluster creation when cluster name length is more than 54", func() {
			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster-234567-foobarfoobar-foobar-fooobaaar-fooo-baaaar"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			properties = {
				rosa_creator_arn = "arn:aws:iam::123456789012:user/dummy",
   				cluster_name = "too_long"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute name string length must be at most 54, got: 59")
		})

		Context("Test destroy cluster", func() {
			BeforeEach(func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage1),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						VerifyJQ(`.name`, "my-cluster"),
						VerifyJQ(`.cloud_provider.id`, "aws"),
						VerifyJQ(`.region.id`, "us-west-1"),
						VerifyJQ(`.product.id`, "rosa"),
						VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, ""),
						VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""),
						VerifyJQ(`.aws.sts.operator_role_prefix`, "test"),
						VerifyJQ(`.aws.sts.role_arn`, ""),
						VerifyJQ(`.aws.sts.support_role_arn`, ""),
						VerifyJQ(`.aws.account_id`, "123456789012"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					}]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, templateReadyState),
					),
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, templateReadyState),
					),
				)
			})

			It("Disable waiting in destroy resource", func() {
				Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					disable_waiting_in_destroy = true
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

				// it should return a warning so exit code will be "0":
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				Expect(Terraform.Destroy().ExitCode).To(BeZero())

			})

			It("Wait in destroy resource but use the default timeout", func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusNotFound, template),
					),
				)
				Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

				// it should return a warning so exit code will be "0":
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				Expect(Terraform.Destroy().ExitCode).To(BeZero())
			})

			It("Wait in destroy resource and set timeout to a negative value", func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusNotFound, template),
					),
				)
				Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					destroy_timeout = -1
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

				// it should return a warning so exit code will be "0":
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				Expect(Terraform.Destroy().ExitCode).To(BeZero())
			})

			It("Wait in destroy resource and set timeout to a positive value", func() {
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusNotFound, template),
					),
				)
				Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					destroy_timeout = 10
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

				// it should return a warning so exit code will be "0":
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				Expect(Terraform.Destroy().ExitCode).To(BeZero())
			})
		})

		It("Disable workload monitor and update it", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.disable_user_workload_monitoring`, true),
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
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : true
					  }
					}]`),
				),
			)
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					disable_workload_monitoring = true
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			// it should return a warning so exit code will be "0":
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// apply for update the workload monitor to be enabled
			// Prepare the server:
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
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : "true"
					  }
					},
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
                        "availability_zones": ["us-west-1a"],
						"compute_machine_type": {
							"id": "r5.xlarge"
						}
					  }
					}]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
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
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : "false"
					  }
					}]`),
				),
			)

			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					disable_workload_monitoring = false
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.disable_workload_monitoring`, false))

		})

		Context("Test Proxy", func() {
			It("Creates cluster with http proxy and update it", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage1),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						VerifyJQ(`.name`, "my-cluster"),
						VerifyJQ(`.cloud_provider.id`, "aws"),
						VerifyJQ(`.region.id`, "us-west-1"),
						VerifyJQ(`.product.id`, "rosa"),
						VerifyJQ(`.proxy.http_proxy`, "http://proxy.com"),
						VerifyJQ(`.proxy.https_proxy`, "https://proxy.com"),
						VerifyJQ(`.additional_trust_bundle`, "123"),
						RespondWithPatchedJSON(http.StatusOK, templateWithTrustBundle, `[
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
					  "path": "/proxy",
					  "value": {
						  "http_proxy" : "http://proxy.com",
						  "https_proxy" : "https://proxy.com"
					  }
					}]`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			proxy = {
				http_proxy = "http://proxy.com",
				https_proxy = "https://proxy.com",
				additional_trust_bundle = "123",
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				// apply for update the proxy's attributes
				// Prepare the server:
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
					  "path": "/proxy",
					  "value": {
						  "http_proxy" : "http://proxy.com",
						  "https_proxy" : "https://proxy.com"
					  }
					},
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "additional_trust_bundle" : "REDUCTED"
					  }
					}]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
						VerifyJQ(`.proxy.https_proxy`, "https://proxy2.com"),
						VerifyJQ(`.proxy.no_proxy`, "test"),
						VerifyJQ(`.additional_trust_bundle`, "123"),
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
					  "path": "/proxy",
					  "value": {
						  "https_proxy" : "https://proxy2.com",
						  "no_proxy" : "test"
					  }
					},
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "additional_trust_bundle" : "REDUCTED"
					  }
					}]`),
					),
				)

				// update the attribute "proxy"
				Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			proxy = {
				https_proxy = "https://proxy2.com",
				no_proxy = "test"
				additional_trust_bundle = "123",
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
				runOutput = Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
				resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
				Expect(resource).To(MatchJQ(`.attributes.proxy.https_proxy`, "https://proxy2.com"))
				Expect(resource).To(MatchJQ(`.attributes.proxy.no_proxy`, "test"))
				Expect(resource).To(MatchJQ(`.attributes.proxy.additional_trust_bundle`, "123"))
			})

			It("Creates cluster without http proxy and update trust bundle - should successes", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage1),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						VerifyJQ(`.name`, "my-cluster"),
						VerifyJQ(`.cloud_provider.id`, "aws"),
						VerifyJQ(`.region.id`, "us-west-1"),
						VerifyJQ(`.product.id`, "rosa"),
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
					}]`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				// apply for update the proxy's attributes
				// Prepare the server:
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
					}]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
						VerifyJQ(`.additional_trust_bundle`, "123"),
						RespondWithPatchedJSON(http.StatusCreated, templateWithTrustBundle, `[
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
					}]`),
					),
				)
				// update the attribute "proxy"
				Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			proxy = {
				additional_trust_bundle = "123",
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
				runOutput = Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})
		})
		It("Creates cluster with default_mp_labels", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.nodes.compute_labels.label_key1`, "label_value1"),
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
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
                        "compute_labels": {
                            "label_key1": "label_value1"
                        },
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    				}
					  }
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
            default_mp_labels = {
                label_key1 = "label_value1"
            }
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.default_mp_labels.label_key1`, "label_value1"))
		})

		It("Accepts to reset proxy values", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.additional_trust_bundle`, "123"),
					RespondWithPatchedJSON(http.StatusOK, templateWithTrustBundle, `[
					{
						"op": "add",
						"path": "/aws",
						"value": {
							"ec2_metadata_http_tokens": "optional",
							"sts": {
								"oidc_endpoint_url": "https://127.0.0.1",
								"thumbprint": "111111",
								"role_arn": "",
								"support_role_arn": "",
								"instance_iam_roles": {
									"master_role_arn": "",
									"worker_role_arn": ""
								},
								"operator_role_prefix": "test"
							}
						}
					}
				]`),
				),
			)

			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			proxy = {
				additional_trust_bundle = "123",
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, templateWithTrustBundle, `[
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
							"worker_role_arn" : ""
						  },
						  "operator_role_prefix" : "test"
					  }
				  }
				}]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
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
							"worker_role_arn" : ""
						  },
						  "operator_role_prefix" : "test"
					  }
				  }
				}]`),
				),
			)

			Terraform.Source(`
			 resource "rhcs_cluster_rosa_classic" "my_cluster" {
			   name           = "my-cluster"
			   cloud_region   = "us-west-1"
				aws_account_id = "123456789012"
				proxy = {
					additional_trust_bundle = ""
				}
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
			 }
			`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

		})
		It("Creates private cluster with aws subnet ids without private link", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.aws.subnet_ids.[0]`, "id1"),
					VerifyJQ(`.aws.private_link`, false),
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"),
					VerifyJQ(`.api.listening`, "internal"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": false,
						  "subnet_ids": ["id1", "id2", "id3"],
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
					  "path": "/api",
					  "value": {
					  	"listening": "internal"
					  }
					},
					{
						"op": "add",
						"path": "/availability_zones",
						"value": ["us-west-1a"]
					},
					{
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    				}
					  }
					}
					]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			availability_zones = ["us-west-1a"]
			aws_private_link = false
			private = true
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
		It("Creates private cluster with aws subnet ids & private link", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.aws.subnet_ids.[0]`, "id1"),
					VerifyJQ(`.aws.private_link`, true),
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"),
					VerifyJQ(`.api.listening`, "internal"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": true,
						  "subnet_ids": ["id1", "id2", "id3"],
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
					  "path": "/api",
					  "value": {
					  	"listening": "internal"
					  }
					},
					{
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    				}
					  }
					}
					]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			availability_zones = ["us-west-1a"]
			private = true
			aws_private_link = true
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Creates cluster when private link is false", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.aws.private_link`, false),
					VerifyJQ(`.api.listening`, "external"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": false,
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			aws_private_link = false
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Creates cluster with shared VPC", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.dns.base_domain`, "mydomain.openshift.dev"),
					VerifyJQ(`.aws.subnet_ids.[0]`, "id1"),
					VerifyJQ(`.aws.private_hosted_zone_id`, "1234"),
					VerifyJQ(`.aws.private_hosted_zone_role_arn`, "arn:aws:iam::111111111111:role/test-shared-vpc"),
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "subnet_ids": ["id1", "id2", "id3"],
						  "ec2_metadata_http_tokens": "optional",
                          "private_hosted_zone_id": "1234",
                          "private_hosted_zone_role_arn": "arn:aws:iam::111111111111:role/test-shared-vpc",
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
						"path": "/dns",
                        "value": {"base_domain": "mydomain.openshift.dev"}
					},
					{
						"op": "add",
						"path": "/availability_zones",
						"value": ["us-west-1a"]
					},
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
            "availability_zones": ["us-west-1a"],
						"compute_machine_type": {
							"id": "r5.xlarge"
						}
					  }
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			availability_zones = ["us-west-1a"]
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
            private_hosted_zone = {
                id = "1234"
                role_arn = "arn:aws:iam::111111111111:role/test-shared-vpc"
            }
            base_dns_domain = "mydomain.openshift.dev"
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Creates rosa sts cluster with autoscaling", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.aws.sts.role_arn`, ""),
					VerifyJQ(`.aws.sts.support_role_arn`, ""),
					VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, ""),
					VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""),
					VerifyJQ(`.aws.sts.operator_role_prefix`, "terraform-operator"),
					VerifyJQ(`.nodes.autoscale_compute.kind`, "MachinePoolAutoscaling"),
					VerifyJQ(`.nodes.autoscale_compute.max_replicas`, float64(4)),
					VerifyJQ(`.nodes.autoscale_compute.min_replicas`, float64(2)),
					VerifyJQ(`.nodes.compute_labels.label_key1`, "label_value1"),
					VerifyJQ(`.nodes.compute_labels.label_key2`, "label_value2"),
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
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
					},
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"autoscale_compute": {
							"min_replicas": 2,
							"max_replicas": 4
						},
						"compute_machine_type": {
							"id": "r5.xlarge"
						},
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
						},
						"availability_zones": [
							"az"
						]
					  }
					}
				  ]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			autoscaling_enabled = "true"
			min_replicas = "2"
			max_replicas = "4"
			default_mp_labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			sts = {
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
				  master_role_arn = "",
				  worker_role_arn = ""
				},
				"operator_role_prefix" : "terraform-operator"
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.autoscaling_enabled`, true))
			Expect(resource).To(MatchJQ(`.attributes.min_replicas`, 2.0))
			Expect(resource).To(MatchJQ(`.attributes.max_replicas`, 4.0))
		})

		It("Creates rosa sts cluster with OIDC Configuration ID", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.aws.sts.role_arn`, ""),
					VerifyJQ(`.aws.sts.support_role_arn`, ""),
					VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, ""),
					VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""),
					VerifyJQ(`.aws.sts.operator_role_prefix`, "terraform-operator"),
					VerifyJQ(`.aws.sts.oidc_config.id`, "aaa"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.1",
							  "oidc_config": {
								"id": "aaa",
								"secret_arn": "aaa",
								"issuer_url": "https://127.0.0.1",
								"reusable": true,
								"managed": false
							  },
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
					}]`),
				),
			)
			// Run the apply command:
			Terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			sts = {
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
				  master_role_arn = "",
				  worker_role_arn = ""
				},
				"operator_role_prefix" : "terraform-operator",
				"oidc_config_id" = "aaa"
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Fails to create cluster with 'openshift-v' prefix in version", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.1",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::765374464689:role/terr-account-Installer-Role",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			version = "openshift-v4.14.0"
			sts = {
				operator_role_prefix = "test"
				role_arn = "arn:aws:iam::765374464689:role/terr-account-Installer-Role",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			// expect to get an error
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`Openshift version must be provided without the "openshift-v" prefix`)
		})

		It("Create cluster with http token", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.aws.ec2_metadata_http_tokens`, "required"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens" : "required",
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			ec2_metadata_http_tokens = "required"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Fails to create cluster with http tokens and not supported version", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.aws.ec2_metadata_http_tokens`, "required"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens" : "required",
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			ec2_metadata_http_tokens = "required"
			version = "4.10.1"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			// expect to get an error
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`Can't build cluster with name 'my-cluster': version '4.10.1' is not supported with ec2_metadata_http_tokens, minimum supported version is 4.11.0`)
		})

		It("Fails to create cluster with http tokens with not supported value", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.aws.http_tokens_state`, "bad_string"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens" : "bad_string",
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
					}]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			ec2_metadata_http_tokens = "bad_string"
			version = "4.12"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			// expect to get an error
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Expected a valid param. Options are [optional required]. Got bad_string.")
		})

		It("Creates cluster with aws additional compute security group ids", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.aws.subnet_ids.[0]`, "id1"),
					VerifyJQ(`.aws.private_link`, false),
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"),
					VerifyJQ(`.api.listening`, "internal"),
					VerifyJQ(`.aws.additional_compute_security_group_ids.[0]`, "id1"),
					VerifyJQ(`.aws.additional_infra_security_group_ids.[0]`, "id2"),
					VerifyJQ(`.aws.additional_control_plane_security_group_ids.[0]`, "id3"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": false,
						  "subnet_ids": ["id1", "id2", "id3"],
						  "additional_compute_security_group_ids": ["id1"],
						  "additional_infra_security_group_ids": ["id2"],
						  "additional_control_plane_security_group_ids": ["id3"],
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
					  "path": "/api",
					  "value": {
					  	"listening": "internal"
					  }
					},
					{
						"op": "add",
						"path": "/availability_zones",
						"value": ["us-west-1a"]
					},
					{
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    				}
					  }
					}
					]`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123456789012"
			availability_zones = ["us-west-1a"]
			aws_private_link = false
			private = true
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
			aws_additional_compute_security_group_ids = [
				"id1"
			]
			aws_additional_infra_security_group_ids = [
				"id2"
			]
			aws_additional_control_plane_security_group_ids = [
				"id3"
			]
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
		  }
		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			// Verify initial cluster version
			resource := Terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.aws_additional_compute_security_group_ids.[0]", "id1"))
			Expect(resource).To(MatchJQ(".attributes.aws_additional_infra_security_group_ids.[0]", "id2"))
			Expect(resource).To(MatchJQ(".attributes.aws_additional_control_plane_security_group_ids.[0]", "id3"))
		})
	})
	Context("rhcs_cluster_rosa_classic - update attributes", func() {
		BeforeEach(func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage1),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.disable_user_workload_monitoring`, true),
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
					  "path": "/",
					  "value": {
						  "external_id" : "123"
					  }
					},
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : true
					  }
					},
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
                        "availability_zones": ["us-west-1a"],
						"compute_machine_type": {
							"id": "r5.xlarge"
						}
					  }
					}]`),
				),
			)
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					disable_workload_monitoring = true
					admin_credentials = {
						username = "cluster_admin"
						password = "1234AbB2341234"
					}
					replicas = "3"
					compute_machine_type = "r5.xlarge"
					tags = {
						"k1" = "v1"
					}
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			// it should return a warning so exit code will be "0":
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

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
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : "true"
					  }
					},
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "external_id" : "123"
					  }
					},
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
                        "availability_zones": ["us-west-1a"],
						"compute_machine_type": {
							"id": "r5.xlarge"
						}
					  }
					}]`),
				),
			)
		})
		It("update admin_credentials, tags, availability_zones and name, and fail", func() {
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					# >>>> update cluster name from "my-cluster" to "my-luster"
					name           = "my-luster"

					replicas = "3"
					compute_machine_type = "r5.xlarge"
					cloud_region   = "us-west-1"
					aws_account_id = "123456789012"
					disable_workload_monitoring = false

					# >>>> update username 
					admin_credentials = {
						username = "cluter_admin"
						password = "1234AbB2341234"
					}
					#external_id           = ""

					# >>>> change the availability zone from "us-west-1a" to "us-west-1b" :
					availability_zones = ["us-west-1b"]

					# >>>> remove tags:
					#tags = {
					#	"k1" = "v1"
					#}
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute admin_credentials, cannot be changed")
		})
		It("update default machine-pool's attributes: autoscaling_enabled", func() {
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					replicas = "3"
					compute_machine_type = "r5.xlarge"

					# >>>> change autoscaling_enabled from null to true
					autoscaling_enabled = true
					aws_account_id = "123456789012"
					disable_workload_monitoring = true
					admin_credentials = {
						username = "cluster_admin"
						password = "1234AbB2341234"
					}
					tags = {
						"k1" = "v1"
					}
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute autoscaling_enabled, cannot be changed from <null> to true")
		})
		It("update default machine-pool's attributes: replicas", func() {
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"

					# >>>> change replicas from 3 to 4
					replicas = "4"

					compute_machine_type = "r5.xlarge"
					aws_account_id = "123456789012"
					disable_workload_monitoring = true
					admin_credentials = {
						username = "cluster_admin"
						password = "1234AbB2341234"
					}
					tags = {
						"k1" = "v1"
					}
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute replicas, cannot be changed from 3 to 4")
		})
		It("update default machine-pool's attributes: default_mp_labels", func() {
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"

					# >>>> change default machine pool labels 
					default_mp_labels = {
						"label_key1" = "label_value1",
						"label_key2" = "label_value2"
					}

					replicas = "3"
					compute_machine_type = "r5.xlarge"
					aws_account_id = "123456789012"
					disable_workload_monitoring = true
					admin_credentials = {
						username = "cluster_admin"
						password = "1234AbB2341234"
					}
					tags = {
						"k1" = "v1"
					}
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute default_mp_labels, cannot be changed")
		})
		It("update default machine-pool's attributes: worker_disk_size", func() {
			Terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"

					# >>>> change worker_disk_size
					worker_disk_size = "2"
					replicas = "3"
					compute_machine_type = "r5.xlarge"
					aws_account_id = "123456789012"
					disable_workload_monitoring = true
					admin_credentials = {
						username = "cluster_admin"
						password = "1234AbB2341234"
					}
					tags = {
						"k1" = "v1"
					}
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
						}
					}
				  }
			`)

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute worker_disk_size, cannot be changed")
		})
	})
})
