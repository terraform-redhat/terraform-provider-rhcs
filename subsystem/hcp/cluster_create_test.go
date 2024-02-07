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
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	. "github.com/onsi/gomega/ghttp"       // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
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

var _ = Describe("rhcs_cluster_rosa_hcp - create", func() {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	baseSpecBuilder := cmv1.NewCluster().
		ID("123").
		ExternalID("123").
		Name("my-cluster").
		AWS(cmv1.NewAWS().
			AccountID("123").
			BillingAccountID("123").
			SubnetIDs("id1", "id2", "id3")).
		State(cmv1.ClusterStateReady).
		Region(cmv1.NewCloudRegion().ID("us-west-1")).
		MultiAZ(true).
		Hypershift(cmv1.NewHypershift().Enabled(true)).
		API(cmv1.NewClusterAPI().URL("https://my-api.example.com")).
		Console(cmv1.NewClusterConsole().URL("https://my-console.example.com")).
		Properties(map[string]string{
			"rosa_tf_version": build.Version,
			"rosa_tf_commit":  build.Commit,
		}).
		Nodes(cmv1.NewClusterNodes().
			Compute(3).AvailabilityZones("us-west-1a").
			ComputeMachineType(cmv1.NewMachineType().ID("r5.xlarge")),
		).
		Network(cmv1.NewNetwork().
			MachineCIDR("10.0.0.0/16").
			ServiceCIDR("172.30.0.0/16").
			PodCIDR("10.128.0.0/14").
			HostPrefix(23)).
		Version(cmv1.NewVersion().ID("openshift-4.8.0")).
		DNS(cmv1.NewDNS().BaseDomain("mycluster-api.example.com"))
	spec, err := baseSpecBuilder.Build()
	Expect(err).ToNot(HaveOccurred())

	b := new(strings.Builder)
	err = cmv1.MarshalCluster(spec, b)
	Expect(err).ToNot(HaveOccurred())
	template := string(b.String())

	baseSpecBuilder.AdditionalTrustBundle("REDACTED")
	specWithTrustBundle, err := baseSpecBuilder.Build()
	Expect(err).ToNot(HaveOccurred())
	b = new(strings.Builder)
	err = cmv1.MarshalCluster(specWithTrustBundle, b)
	Expect(err).ToNot(HaveOccurred())
	templateWithTrustBundle := string(b.String())

	Context("rhcs_cluster_rosa_hcp - create", func() {
		It("invalid az for region", func() {
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  aws_billing_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  worker_role_arn = "",
				  }
			  }
			  aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
			  version = "openshift-v4.11.1"
			}
		  `)
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		It("version with unsupported prefix error", func() {
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
              availability_zones = ["us-east-1a"]
			  aws_account_id = "123"
			  aws_billing_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  worker_role_arn = "",
				  }
			  }
			  aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.2",
								  "thumbprint": "111111",
								  "role_arn": "",
								  "support_role_arn": "",
								  "instance_iam_roles" : {
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
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  aws_billing_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  worker_role_arn = "",
				  }
			  }
			  aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.2",
								  "thumbprint": "111111",
								  "role_arn": "",
								  "support_role_arn": "",
								  "instance_iam_roles" : {
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
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  aws_billing_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  worker_role_arn = "",
				  }
			  }
			  aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  aws_billing_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  worker_role_arn = "",
				  }
			  }
			  aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
			wait_for_create_complete = true
		  }
		`)
				Expect(terraform.Apply()).ToNot(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
			wait_for_create_complete = true
		  }
		`)
				Expect(terraform.Apply()).To(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					VerifyJQ(`.properties.`+prop_key, prop_val),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `"` +
				`}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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

					VerifyJQ(`.properties.`+prop_key, prop_val),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `"` +
				`}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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

					VerifyJQ(`.properties.`+prop_key, prop_val+"_1"),
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `_1"` +
				`}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					VerifyJQ(`.properties.`+prop_key, prop_val),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `"` +
				`}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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

					VerifyJQ(`.properties.`+prop_key, nil),
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
            properties = {}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
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

					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			properties = {
   				rosa_tf_version = "bob"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Should fail cluster creation when cluster name length is more than 15", func() {
			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster-234567"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			properties = {
   				cluster_name = "too_long"
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, template),
					),
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, template),
					),
				)
			})

			It("Disable waiting in destroy resource", func() {
				terraform.Source(`
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					disable_waiting_in_destroy = true
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					destroy_timeout = -1
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					destroy_timeout = 10
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
					}]`),
				),
			)
			terraform.Source(`
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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

			terraform.Source(`
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
				  }
			`)

			Expect(terraform.Apply()).To(BeZero())

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
						RespondWithPatchedJSON(http.StatusOK, templateWithTrustBundle, `[
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
				terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
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
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
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
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
				Expect(terraform.Apply()).To(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
				Expect(resource).To(MatchJQ(`.attributes.proxy.https_proxy`, "https://proxy2.com"))
				Expect(resource).To(MatchJQ(`.attributes.proxy.no_proxy`, "test"))
				Expect(resource).To(MatchJQ(`.attributes.proxy.additional_trust_bundle`, "123"))
			})

			It("Creates cluster without http proxy and update trust bundle - should successes", func() {
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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

				// Run the apply command:
				terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						VerifyJQ(`.additional_trust_bundle`, "123"),
						RespondWithPatchedJSON(http.StatusCreated, templateWithTrustBundle, `[
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
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
					}]`),
					),
				)
				// update the attribute "proxy"
				terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			proxy = {
				additional_trust_bundle = "123",
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
				Expect(terraform.Apply()).To(BeZero())
			})
		})

		It("Except to fail on proxy validators", func() {
			// Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle
			terraform.Source(`
			 resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			   name           = "my-cluster"
			   cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				proxy = {
				}
				sts = {
					operator_role_prefix = "test"
					role_arn = "",
					support_role_arn = "",
					instance_iam_roles = {
						worker_role_arn = "",
					}
				}
				aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
			 }
			`)
			Expect(terraform.Apply()).NotTo(BeZero())

			// Prepare the server:
			server.AppendHandlers(
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
							"sts": {
								"oidc_endpoint_url": "https://127.0.0.2",
								"thumbprint": "111111",
								"role_arn": "",
								"support_role_arn": "",
								"instance_iam_roles": {
									"worker_role_arn": ""
								},
								"operator_role_prefix": "test"
							}
						}
					}
				]`),
				),
			)

			terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			proxy = {
				additional_trust_bundle = "123",
			}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			availability_zones = ["us-west-1a"]
			aws_private_link = true
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			aws_private_link = false
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
					VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""),
					VerifyJQ(`.aws.sts.operator_role_prefix`, "terraform-operator"),
					VerifyJQ(`.aws.sts.oidc_config.id`, "aaa"),
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
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
		resource "rhcs_cluster_rosa_hcp" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			sts = {
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
				  worker_role_arn = ""
				},
				"operator_role_prefix" : "terraform-operator",
				"oidc_config_id" = "aaa"
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::765374464689:role/terr-account-Installer-Role",
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

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_billing_account_id = "123"
			version = "openshift-v4.12"
			sts = {
				operator_role_prefix = "test"
				role_arn = "arn:aws:iam::765374464689:role/terr-account-Installer-Role",
				support_role_arn = "",
				instance_iam_roles = {
					worker_role_arn = "",
				}
			}
			aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
		  }
		`)
			// expect to get an error
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		Context("rhcs_cluster_rosa_hcp - update attributes", func() {
			BeforeEach(func() {
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
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
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
				terraform.Source(`
				  resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
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
							worker_role_arn = "",
						}
					}
					aws_subnet_ids = [
				"id1", "id2", "id3"
			  ]
				  }
			`)

				// it should return a warning so exit code will be "0":
				Expect(terraform.Apply()).To(BeZero())

				server.AppendHandlers(
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
		})
	})
})
