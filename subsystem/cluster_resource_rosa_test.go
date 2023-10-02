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
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
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
	  "name": "my-cluster",
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
			terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
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
			}
		  `)
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		It("version with unsupported prefix error", func() {
			terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
              availability_zones = ["us-east-1a"]
			  aws_account_id = "123"
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
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		Context("Test channel groups", func() {
			It("doesn't append the channel group when on the default channel", func() {
				server.AppendHandlers(
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
								  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
			})
			It("appends the channel group when on a non-default channel", func() {
				server.AppendHandlers(
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
								  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
			})
			It("returns an error when the version is not found in the channel group", func() {
				server.AppendHandlers(
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
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
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
				Expect(terraform.Apply()).NotTo(BeZero())
			})
		})

		Context("Test wait attribute", func() {
			It("Create cluster and wait till it will be in error state", func() {
				// Prepare the server:
				server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
				Expect(terraform.Apply()).ToNot(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.state", "error"))
			})

			It("Create cluster and wait till it will be in ready state", func() {
				// Prepare the server:
				server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.state", "ready"))
			})
		})
		It("Creates basic cluster", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
		})

		It("Creates basic cluster returned empty az list", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
		})

		It("Creates basic cluster with admin user", func() {
			// Prepare the server:
			server.AppendHandlers(
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
					VerifyJQ(`.htpasswd.users.items[0].username`, "cluster_admin"),
					VerifyJQ(`.htpasswd.users.items[0].password`, "1234AbB2341234"),
					VerifyJQ(`.properties.rosa_tf_version`, build.Version),
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            admin_credentials = {
                username = "cluster_admin"
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
		})

		It("Creates basic cluster - and reconcile on a 404", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
			Expect(resource).To(MatchJQ(".attributes.id", "123")) // cluster has id 123

			// Prepare the server for reconcile
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "openshift-4.8.0"))
			Expect(resource).To(MatchJQ(".attributes.id", "1234")) // reconciled cluster has id of 1234
		})

		It("Creates basic cluster with properties", func() {
			prop_key := "my_prop_key"
			prop_val := "my_prop_val"
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `"` +
				`}
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates basic cluster with properties and update them", func() {
			prop_key := "my_prop_key"
			prop_val := "my_prop_val"
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `"` +
				`}
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+prop_key, prop_val))

			// Prepare server for update
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`_1"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `_1"` +
				`}
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
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val+"_1"))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+prop_key, prop_val+"_1"))
		})

		It("Creates basic cluster with properties and delete them", func() {
			prop_key := "my_prop_key"
			prop_val := "my_prop_val"
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`),
				),
			)

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `"` +
				`}
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+prop_key, prop_val))
			Expect(resource).To(MatchJQ(`.attributes.properties| keys | length`, 1))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties| keys | length`, 3))

			// Prepare server for update
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = {}
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
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.properties | keys | length`, 0))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties | keys | length`, 2))
		})

		It("Should fail cluster creation when trying to override reserved properties", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			properties = {
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Should fail cluster creation when cluster name length is more than 15", func() {
			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster-234567"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			properties = {
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
			Expect(terraform.Apply()).ToNot(BeZero())

		})

		Context("Test destroy cluster", func() {
			BeforeEach(func() {
				server.AppendHandlers(
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
						VerifyJQ(`.aws.account_id`, "123"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
					      "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())

			})

			It("Wait in destroy resource but use the default timeout", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusNotFound, template),
					),
				)
				terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})

			It("Wait in destroy resource and set timeout to a negative value", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusNotFound, template),
					),
				)
				terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})

			It("Wait in destroy resource and set timeout to a positive value", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusNotFound, template),
					),
				)
				terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})
		})

		It("Disable workload monitor and update it", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())

			// apply for update the workload monitor to be enabled
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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

			terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
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

			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.disable_workload_monitoring`, false))

		})

		Context("Test Proxy", func() {
			It("Creates cluster with http proxy and update it", func() {
				// Prepare the server:
				server.AppendHandlers(
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
						RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						  "additional_trust_bundle" : "123"
					  }
					}]`),
					),
				)

				// Run the apply command:
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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

				Expect(terraform.Apply()).To(BeZero())

				// apply for update the proxy's attributes
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
				Expect(terraform.Apply()).To(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
				Expect(resource).To(MatchJQ(`.attributes.proxy.https_proxy`, "https://proxy2.com"))
				Expect(resource).To(MatchJQ(`.attributes.proxy.no_proxy`, "test"))
				Expect(resource).To(MatchJQ(`.attributes.proxy.additional_trust_bundle`, "123"))
			})
			It("Creates cluster without http proxy and update trust bundle - should fail", func() {
				// Prepare the server:
				server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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

				Expect(terraform.Apply()).To(BeZero())

				// apply for update the proxy's attributes
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						  "additional_trust_bundle" : "123"
					  }
					}]`),
					),
				)
				// update the attribute "proxy"
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
				Expect(terraform.Apply()).ToNot(BeZero())
			})
		})
		It("Creates cluster with default_mp_labels and update them", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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

			Expect(terraform.Apply()).To(BeZero())

			// apply for update the default_mp_labels
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				// Update handler and response
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.nodes.compute_labels.changed_label`, "changed"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                            "changed_label": "changed"
                        },
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

			// update the attribute "proxy"
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            default_mp_labels = {
                changed_label = "changed"
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
			Expect(terraform.Apply()).To(BeZero(), "Failed to update cluster with changed default_mp_labels")
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.default_mp_labels.changed_label`, "changed"))
		})

		It("Except to fail on proxy validators", func() {
			// Expected at least one of the following: http-proxy, https-proxy
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			proxy = {
				no_proxy = "test1, test2"
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
			Expect(terraform.Apply()).NotTo(BeZero())

			// Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle
			terraform.Source(`
			 resource "rhcs_cluster_rosa_classic" "my_cluster" {
			   name           = "my-cluster"
			   cloud_region   = "us-west-1"
				aws_account_id = "123"
				proxy = {
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
			Expect(terraform.Apply()).NotTo(BeZero())
		})
		It("Creates private cluster with aws subnet ids without private link", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
		})
		It("Creates private cluster with aws subnet ids & private link", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates cluster when private link is false", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates cluster with shared VPC", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates rosa sts cluster with autoscaling and update the default machine pool", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())

			// apply for update the min_replica from 2 to 3
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
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
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.nodes.autoscale_compute.kind`, "MachinePoolAutoscaling"),
					VerifyJQ(`.nodes.autoscale_compute.max_replicas`, float64(4)),
					VerifyJQ(`.nodes.autoscale_compute.min_replicas`, float64(3)),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
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
							"min_replicas": 3,
							"max_replicas": 4
						},
						"compute_machine_type": {
							"id": "r5.xlarge"
						},
						"availability_zones": ["az"],
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
						}
					  }
					}
				  ]`),
				),
			)
			// Run the apply command:
			terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			autoscaling_enabled = "true"
			min_replicas = "3"
			max_replicas = "4"
			default_mp_labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			sts = {
				role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
				support_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
				instance_iam_roles = {
				  master_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
				  worker_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
				},
				"operator_role_prefix" : "terraform-operator"
			}
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())

			// apply for update the autoscaling group to compute nodes
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
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
							"min_replicas": 3,
							"max_replicas": 4
						},
						"availability_zones": ["az"],
						"compute_machine_type": {
							"id": "r5.xlarge"
						},
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
						}
					  }
					}
				  ]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(`.nodes.compute`, float64(4)),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
					},
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 4,
						"availability_zones": ["az"],
						"compute_machine_type": {
							"id": "r5.xlarge"
						},
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
						}
					  }
					}
				  ]`),
				),
			)
			// Run the apply command:
			terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			replicas = 4
			default_mp_labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
			}
			sts = {
				role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
				support_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
				instance_iam_roles = {
				  master_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
				  worker_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
				},
				"operator_role_prefix" : "terraform-operator"
			}
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates rosa sts cluster with OIDC Configuration ID", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "oidc_config": {
								"id": "aaa",
								"secret_arn": "aaa",
								"issuer_url": "https://127.0.0.2",
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
			terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Fails to create cluster with incompatible account role's version and fail", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			version = "openshift-v4.12"
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Create cluster with http token", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Fails to create cluster with http tokens and not supported version", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			ec2_metadata_http_tokens = "required"
			version = "openshift-v4.10"
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Fails to create cluster with http tokens with not supported value", func() {
			// Prepare the server:
			server.AppendHandlers(
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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})
	})
})

var _ = Describe("rhcs_cluster_rosa_classic - upgrade", func() {
	const template = `{
		"id": "123",
		"name": "my-cluster",
		"state": "ready",
		"region": {
		  "id": "us-west-1"
		},
		"aws": {
			"ec2_metadata_http_tokens": "optional",
			"sts": {
				"oidc_endpoint_url": "https://127.0.0.2",
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
			"compute": 3,
	        "availability_zones": ["az"],
			"compute_machine_type": {
				"id": "r5.xlarge"
			}
		},
		"version": {
			"id": "4.10.0"
		}
	}`
	const versionList = `{
		"kind": "VersionList",
		"page": 1,
		"size": 3,
		"total": 3,
		"items": [{
				"kind": "Version",
				"id": "openshift-v4.10.0",
				"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.0",
				"raw_id": "4.10.0"
			},
			{
				"kind": "Version",
				"id": "openshift-v4.10.1",
				"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.1",
				"raw_id": "4.10.1"
			},
			{
				"kind": "Version",
				"id": "openshift-v4.10.1",
				"href": "/api/clusters_mgmt/v1/versions/openshift-v4.11.1",
				"raw_id": "4.11.1"
			}
		]
	}`
	const v4_10_0Info = `{
		"kind": "Version",
		"id": "openshift-v4.10.0",
		"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.0",
		"raw_id": "4.10.0",
		"enabled": true,
		"default": false,
		"channel_group": "stable",
		"available_upgrades": ["4.10.1"],
		"rosa_enabled": true
	}`
	const v4_10_1Info = `{
		"kind": "Version",
		"id": "openshift-v4.10.1",
		"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.1",
		"raw_id": "4.10.1",
		"enabled": true,
		"default": false,
		"channel_group": "stable",
		"available_upgrades": [],
		"rosa_enabled": true
	}`
	const upgradePoliciesEmpty = `{
		"kind": "UpgradePolicyList",
		"page": 1,
		"size": 0,
		"total": 0,
		"items": []
	}`
	BeforeEach(func() {
		Expect(json.Valid([]byte(template))).To(BeTrue())
		Expect(json.Valid([]byte(v4_10_0Info))).To(BeTrue())

		// Create a cluster for us to upgrade:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				RespondWithJSON(http.StatusOK, versionList),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
					}
				]`),
			),
		)
		terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				},
				"operator_iam_roles": [
					{
					  "id": "",
					  "name": "ebs-cloud-credentials",
					  "namespace": "openshift-cluster-csi-drivers",
					  "role_arn": "",
					  "service_account": ""
					},
					{
					  "id": "",
					  "name": "cloud-credentials",
					  "namespace": "openshift-cloud-network-config-controller",
					  "role_arn": "",
					  "service_account": ""
					},
					{
					  "id": "",
					  "name": "aws-cloud-credentials",
					  "namespace": "openshift-machine-api",
					  "role_arn": "",
					  "service_account": ""
					},
					{
					  "id": "",
					  "name": "cloud-credential-operator-iam-ro-creds",
					  "namespace": "openshift-cloud-credential-operator",
					  "role_arn": "",
					  "service_account": ""
					},
					{
					  "id": "",
					  "name": "installer-cloud-credentials",
					  "namespace": "openshift-image-registry",
					  "role_arn": "",
					  "service_account": ""
					},
					{
					  "id": "",
					  "name": "cloud-credentials",
					  "namespace": "openshift-ingress-operator",
					  "role_arn": "",
					  "service_account": ""
					}
				]
			}
			version = "4.10.0"
		}
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Verify initial cluster version
		resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster")
		Expect(resource).To(MatchJQ(".attributes.current_version", "4.10.0"))
	})
	Context("rhcs_cluster_rosa_classic - create", func() {
		It("Upgrades cluster", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"),
					RespondWithJSON(http.StatusOK, v4_10_0Info),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"),
					RespondWithJSON(http.StatusOK, v4_10_1Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					RespondWithJSON(http.StatusOK, upgradePoliciesEmpty),
				),
				// Look for gate agreements by posting an upgrade policy w/ dryRun
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"),
					VerifyJQ(".version", "4.10.1"),
					RespondWithJSON(http.StatusBadRequest, `{
					"kind": "Error",
					"id": "400",
					"href": "/api/clusters_mgmt/v1/errors/400",
					"code": "CLUSTERS-MGMT-400",
					"reason": "There are missing version gate agreements for this cluster. See details.",
					"details": [
					  {
						"kind": "VersionGate",
						"id": "999",
						"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
						"version_raw_id_prefix": "4.10",
						"label": "api.openshift.com/gate-sts",
						"value": "4.10",
						"warning_message": "STS roles must be updated blah blah blah",
						"description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
						"documentation_url": "https://access.redhat.com/solutions/0000000",
						"sts_only": true,
						"creation_timestamp": "2023-04-03T06:39:57.057613Z"
					  }
					],
					"operation_id": "8f2d2946-c4ef-4c2f-877b-c19eb17dc918"
				  }`),
				),
				// Send acks for all gate agreements
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/gate_agreements"),
					VerifyJQ(".version_gate.id", "999"),
					RespondWithJSON(http.StatusCreated, `{
					"kind": "VersionGateAgreement",
					"id": "888",
					"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
					"version_gate": {
					  "kind": "VersionGate",
					  "id": "999",
					  "href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
					  "version_raw_id_prefix": "4.10",
					  "label": "api.openshift.com/gate-sts",
					  "value": "4.10",
					  "warning_message": "STS blah blah blah",
					  "description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
					  "documentation_url": "https://access.redhat.com/solutions/0000000",
					  "sts_only": true,
					  "creation_timestamp": "2023-04-03T06:39:57.057613Z"
					},
					"creation_timestamp": "2023-06-21T13:02:06.291443Z"
				  }`),
				),
				// Create an upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					VerifyJQ(".version", "4.10.1"),
					RespondWithJSON(http.StatusCreated, `
				{
					"kind": "UpgradePolicy",
					"id": "123",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
					"schedule_type": "manual",
					"upgrade_type": "OSD",
					"version": "4.10.1",
					"next_run": "2023-06-09T20:59:00Z",
					"cluster_id": "123",
					"enable_minor_version_upgrades": true
				}`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
						{
						  "op": "add",
						  "path": "/properties",
						  "value": {
							"rosa_tf_commit": "123",
							"rosa_tf_version": "123"
						  }
						}
					]`,
					)),
			)
			// Perform upgrade w/ auto-ack of sts-only gate agreements
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = ""
				support_role_arn = ""
				instance_iam_roles = {
					master_role_arn = ""
					worker_role_arn = ""
				}
			}
			version = "4.10.1"
		}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Upgrades cluster support old version format", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"),
					RespondWithJSON(http.StatusOK, v4_10_0Info),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"),
					RespondWithJSON(http.StatusOK, v4_10_1Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					RespondWithJSON(http.StatusOK, upgradePoliciesEmpty),
				),
				// Look for gate agreements by posting an upgrade policy w/ dryRun
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"),
					VerifyJQ(".version", "4.10.1"),
					RespondWithJSON(http.StatusBadRequest, `{
					"kind": "Error",
					"id": "400",
					"href": "/api/clusters_mgmt/v1/errors/400",
					"code": "CLUSTERS-MGMT-400",
					"reason": "There are missing version gate agreements for this cluster. See details.",
					"details": [
					  {
						"kind": "VersionGate",
						"id": "999",
						"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
						"version_raw_id_prefix": "4.10",
						"label": "api.openshift.com/gate-sts",
						"value": "4.10",
						"warning_message": "STS roles must be updated blah blah blah",
						"description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
						"documentation_url": "https://access.redhat.com/solutions/0000000",
						"sts_only": true,
						"creation_timestamp": "2023-04-03T06:39:57.057613Z"
					  }
					],
					"operation_id": "8f2d2946-c4ef-4c2f-877b-c19eb17dc918"
				  }`),
				),
				// Send acks for all gate agreements
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/gate_agreements"),
					VerifyJQ(".version_gate.id", "999"),
					RespondWithJSON(http.StatusCreated, `{
					"kind": "VersionGateAgreement",
					"id": "888",
					"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
					"version_gate": {
					  "kind": "VersionGate",
					  "id": "999",
					  "href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
					  "version_raw_id_prefix": "4.10",
					  "label": "api.openshift.com/gate-sts",
					  "value": "4.10",
					  "warning_message": "STS blah blah blah",
					  "description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
					  "documentation_url": "https://access.redhat.com/solutions/0000000",
					  "sts_only": true,
					  "creation_timestamp": "2023-04-03T06:39:57.057613Z"
					},
					"creation_timestamp": "2023-06-21T13:02:06.291443Z"
				  }`),
				),
				// Create an upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					VerifyJQ(".version", "4.10.1"),
					RespondWithJSON(http.StatusCreated, `
				{
					"kind": "UpgradePolicy",
					"id": "123",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
					"schedule_type": "manual",
					"upgrade_type": "OSD",
					"version": "4.10.1",
					"next_run": "2023-06-09T20:59:00Z",
					"cluster_id": "123",
					"enable_minor_version_upgrades": true
				}`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
					}
				]`),
				))
			// Perform upgrade w/ auto-ack of sts-only gate agreements
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			version = "openshift-v4.10.1"
		}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Does nothing if upgrade is in progress", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"),
					RespondWithJSON(http.StatusOK, v4_10_0Info),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"),
					RespondWithJSON(http.StatusOK, v4_10_1Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"page": 1,
					"size": 0,
					"total": 0,
					"items": [
						{
							"kind": "UpgradePolicy",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
							"schedule_type": "manual",
							"upgrade_type": "OSD",
							"version": "4.10.0",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
						}
					]
				}`),
				),
				// Check it's state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"id": "456",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state",
					"description": "Upgrade in progress",
					"value": "started"
				}`),
				),
			)
			// Perform try the upgrade
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			version = "4.10.1"
		}`)
			// Will fail due to upgrade in progress
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		It("Cancels and upgrade for the wrong version & schedules new", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"),
					RespondWithJSON(http.StatusOK, v4_10_0Info),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"),
					RespondWithJSON(http.StatusOK, v4_10_1Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"page": 1,
					"size": 0,
					"total": 0,
					"items": [
						{
							"kind": "UpgradePolicy",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
							"schedule_type": "manual",
							"upgrade_type": "OSD",
							"version": "4.10.0",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
						}
					]
				}`),
				),
				// Check it's state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"id": "456",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state",
					"description": "",
					"value": "scheduled"
				}`),
				),
				// Delete existing upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				// Look for gate agreements by posting an upgrade policy w/ dryRun (no gates necessary)
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"),
					VerifyJQ(".version", "4.10.1"),
					RespondWithJSON(http.StatusNoContent, ""),
				),
				// Create an upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					VerifyJQ(".version", "4.10.1"),
					RespondWithJSON(http.StatusCreated, `
				{
					"kind": "UpgradePolicy",
					"id": "123",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
					"schedule_type": "manual",
					"upgrade_type": "OSD",
					"version": "4.10.1",
					"next_run": "2023-06-09T20:59:00Z",
					"cluster_id": "123",
					"enable_minor_version_upgrades": true
				}`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
					}
				]`),
				),
			)
			// Perform try the upgrade
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			version = "4.10.1"
		}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Cancels upgrade if version=current_version", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"page": 1,
					"size": 0,
					"total": 0,
					"items": [
						{
							"kind": "UpgradePolicy",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
							"schedule_type": "manual",
							"upgrade_type": "OSD",
							"version": "4.10.1",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
						}
					]
				}`),
				),
				// Check it's state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"id": "456",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state",
					"description": "",
					"value": "scheduled"
				}`),
				),
				// Delete existing upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
			)
			// Set version to match current cluster version
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			version = "4.10.0"
		}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("is an error to request a version older than current", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template,
						`[
					{
						"op": "replace",
						"path": "/version/id",
						"value": "openshift-v4.11.0"
					}]`),
				),
			)
			// Set version to before current cluster version, but after version from create
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			version = "4.10.1"
		}`)
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		It("older than current is allowed as long as not changed", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithPatchedJSON(http.StatusOK, template,
						`[
						{
							"op": "replace",
							"path": "/version/id",
							"value": "openshift-v4.11.0"
						}]`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, template),
				),
			)
			// Set version to before current cluster version, but matching what was
			// used during creation (i.e. in state file)
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
				}
			}
			version = "4.10.0"
		}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		Context("Un-acked gates", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					// Refresh cluster state
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, template),
					),
					// Validate upgrade versions
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"),
						RespondWithJSON(http.StatusOK, v4_10_0Info),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"),
						RespondWithJSON(http.StatusOK, v4_10_1Info),
					),
					// Look for existing upgrade policies
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
						RespondWithJSON(http.StatusOK, upgradePoliciesEmpty),
					),
					// Look for gate agreements by posting an upgrade policy w/ dryRun
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"),
						VerifyJQ(".version", "4.10.1"),
						RespondWithJSON(http.StatusBadRequest, `{
						"kind": "Error",
						"id": "400",
						"href": "/api/clusters_mgmt/v1/errors/400",
						"code": "CLUSTERS-MGMT-400",
						"reason": "There are missing version gate agreements for this cluster. See details.",
						"details": [
						{
							"kind": "VersionGate",
							"id": "999",
							"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
							"version_raw_id_prefix": "4.10",
							"label": "api.openshift.com/ackme",
							"value": "4.10",
							"warning_message": "user gotta ack",
							"description": "deprecations... blah blah blah",
							"documentation_url": "https://access.redhat.com/solutions/0000000",
							"sts_only": false,
							"creation_timestamp": "2023-04-03T06:39:57.057613Z"
						}
						],
						"operation_id": "8f2d2946-c4ef-4c2f-877b-c19eb17dc918"
					}`),
					),
				)
			})
			It("Fails upgrade for un-acked gates", func() {
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
				version = "openshift-v4.10.1"
			}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
			It("Fails upgrade if wrong version is acked", func() {
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
				version = "openshift-v4.10.1"
				upgrade_acknowledgements_for = "1.1"
			}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
			It("It acks gates if correct ack is provided", func() {
				server.AppendHandlers(
					// Send acks for all gate agreements
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/gate_agreements"),
						VerifyJQ(".version_gate.id", "999"),
						RespondWithJSON(http.StatusCreated, `{
					"kind": "VersionGateAgreement",
					"id": "888",
					"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
					"version_gate": {
					  "kind": "VersionGate",
					  "id": "999",
					  "href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
					  "version_raw_id_prefix": "4.10",
					  "label": "api.openshift.com/gate-sts",
					  "value": "4.10",
					  "warning_message": "blah blah blah",
					  "description": "whatever",
					  "documentation_url": "https://access.redhat.com/solutions/0000000",
					  "sts_only": false,
					  "creation_timestamp": "2023-04-03T06:39:57.057613Z"
					},
					"creation_timestamp": "2023-06-21T13:02:06.291443Z"
				  }`),
					),
					// Create an upgrade policy
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"),
						VerifyJQ(".version", "4.10.1"),
						RespondWithJSON(http.StatusCreated, `
				{
					"kind": "UpgradePolicy",
					"id": "123",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
					"schedule_type": "manual",
					"upgrade_type": "OSD",
					"version": "4.10.1",
					"next_run": "2023-06-09T20:59:00Z",
					"cluster_id": "123",
					"enable_minor_version_upgrades": true
				}`),
					),
					// Patch the cluster (w/ no changes)
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
					}
				]`),
					),
				)
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						master_role_arn = "",
						worker_role_arn = "",
					}
				}
				version = "4.10.1"
				upgrade_acknowledgements_for = "4.10"
			}`)
				Expect(terraform.Apply()).To(BeZero())
			})
		})
	})
})

var _ = Describe("rhcs_cluster_rosa_classic - import", func() {
	const template = `{
		"id": "123",
		"name": "my-cluster",
		"region": {
		  "id": "us-west-1"
		},
		"aws": {
			"sts": {
				"oidc_endpoint_url": "https://127.0.0.2",
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
	Context("rhcs_cluster_rosa_classic - create", func() {
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
								  "oidc_endpoint_url": "https://127.0.0.2",
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
