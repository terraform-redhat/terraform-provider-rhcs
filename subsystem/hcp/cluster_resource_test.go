/*
Copyright (c) 2024 Red Hat, Inc.

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
	"fmt"
	"net/http"
	"strings"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	. "github.com/onsi/gomega/ghttp"       // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

const (
	propKey         = "prop_key"
	propValue       = "prop_value"
	cluster123Route = "/api/clusters_mgmt/v1/clusters/123"
)

var _ = Describe("HCP Cluster", func() {
	// cmv1 doesn't have a marshaling option for page
	versionListPage := `
	{
		"kind": "VersionList",
		"page": 1,
		"size": 2,
		"total": 2,
		"items": [	
			{
				"kind": "Version",
				"id": "openshift-v4.14.0",
				"href": "/api/clusters_mgmt/v1/versions/openshift-v4.14.0",
				"raw_id": "4.14.0"
			},
			{
				"kind": "Version",
				"id": "openshift-v4.14.1",
				"href": "/api/clusters_mgmt/v1/versions/openshift-v4.14.1",
				"raw_id": "4.14.1"
			}
		]
	}`
	v4140Builder := cmv1.NewVersion().ChannelGroup("stable").
		Enabled(true).
		ROSAEnabled(true).
		HostedControlPlaneEnabled(true).
		ID("openshift-v4.14.0").
		RawID("v4.14.0").
		AvailableUpgrades("4.14.1")
	v4141Spec, err := cmv1.NewVersion().ChannelGroup("stable").
		Enabled(true).
		ROSAEnabled(true).
		HostedControlPlaneEnabled(true).
		ID("openshift-v4.14.1").
		RawID("v4.14.1").Build()
	b := new(strings.Builder)
	err = cmv1.MarshalVersion(v4141Spec, b)
	Expect(err).ToNot(HaveOccurred())
	v4141Info := b.String()
	const emptyControlPlaneUpgradePolicies = `
	{
		"page": 1,
		"size": 0,
		"total": 0,
		"items": []
	}`
	Expect(err).NotTo(HaveOccurred())
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
			Compute(3).AvailabilityZones("us-west-1a", "us-west-1b", "us-west-1c").
			ComputeMachineType(cmv1.NewMachineType().ID("r5.xlarge")),
		).
		Network(cmv1.NewNetwork().
			MachineCIDR("10.0.0.0/16").
			ServiceCIDR("172.30.0.0/16").
			PodCIDR("10.128.0.0/14").
			HostPrefix(23)).
		Version(v4140Builder).
		DNS(cmv1.NewDNS().BaseDomain("mycluster-api.example.com"))
	spec, err := baseSpecBuilder.Build()
	Expect(err).ToNot(HaveOccurred())

	b = new(strings.Builder)
	err = cmv1.MarshalCluster(spec, b)
	Expect(err).ToNot(HaveOccurred())
	template := b.String()

	Context("Create/Update/Delete", func() {
		baseSpecBuilder.AdditionalTrustBundle("REDACTED")
		specWithTrustBundle, err := baseSpecBuilder.Build()
		Expect(err).ToNot(HaveOccurred())
		b = new(strings.Builder)
		err = cmv1.MarshalCluster(specWithTrustBundle, b)
		Expect(err).ToNot(HaveOccurred())
		templateWithTrustBundle := b.String()
		Context("Availability zones", func() {
			It("invalid az for region", func() {
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
				  version = "4.14.1"
				}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
		})

		Context("Version", func() {
			It("version with unsupported prefix error", func() {
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
				  version = "openshift-v4.14.1"
				}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
		})
		Context("Test channel groups", func() {
			It("doesn't append the channel group when on the default channel", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						VerifyJQ(`.version.id`, "openshift-v4.14.1"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
						},
						{
							"op": "replace",
							"path": "/version/id",
							"value": "openshift-v4.14.1"
						}]`),
					))
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					version = "4.14.1"
				}`)
				Expect(terraform.Apply()).To(BeZero())
			})
			It("appends the channel group when on a non-default channel", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithPatchedJSON(http.StatusOK, versionListPage, `[
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
						}]`),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					channel_group = "fast"
					version = "4.50.0"
				}`)
				Expect(terraform.Apply()).To(BeZero())
			})
			It("returns an error when the version is not found in the channel group", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithPatchedJSON(http.StatusOK, versionListPage, `[
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
					channel_group = "fast"
					version = "4.99.99"
				}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
		})
		Context("Test wait attribute", func() {
			It("Create cluster and wait till it will be in error state", func() {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
						VerifyRequest(http.MethodGet, cluster123Route),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					wait_for_create_complete = true
				}`)
				Expect(terraform.Apply()).ToNot(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.state", "error"))
			})

			It("Create cluster and wait till it will be in ready state", func() {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
						VerifyRequest(http.MethodGet, cluster123Route),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					wait_for_create_complete = true
				}`)
				Expect(terraform.Apply()).To(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.state", "ready"))
			})
		})
		It("Create cluster and wait til std compute nodes are up", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					RespondWithPatchedJSON(http.StatusCreated, template, `[
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
					VerifyRequest(http.MethodGet, cluster123Route),
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
					},
					{
					"op": "add",
						"path": "/state",
						"value": "ready"
					}]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
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
					},
					{
						"op": "add",
						"path": "/status",
						"value": {
							"current_compute": 3
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				wait_for_create_complete = true
				wait_for_std_compute_nodes_complete = true
			}`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.state", "ready"))
		})
		Context("Create cluster and wait until std compute nodes are up", func() {
			BeforeEach(func() {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
						RespondWithPatchedJSON(http.StatusCreated, template, `[
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
						VerifyRequest(http.MethodGet, cluster123Route),
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
				},
				{
				"op": "add",
					"path": "/state",
					"value": "ready"
				}]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
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
				},
				{
					"op": "add",
					"path": "/status",
					"value": {
						"current_compute": 3
					}
				}]`),
					),
				)
			})
			When("not waiting for creating alongside", func() {
				It("fails - wait_for_create_complete = false", func() {
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				wait_for_create_complete = false
				wait_for_std_compute_nodes_complete = true
			}`)
					Expect(terraform.Apply()).ToNot(BeZero())
				})
				It("fails - wait_for_create_complete = nil", func() {
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				wait_for_std_compute_nodes_complete = true
			}`)
					Expect(terraform.Apply()).ToNot(BeZero())
				})
			})
			When("waiting for creating alongside", func() {
				It("succeeds", func() {
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				wait_for_create_complete = true
				wait_for_std_compute_nodes_complete = true
			}`)
					Expect(terraform.Apply()).To(BeZero())
					resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
					Expect(resource).To(MatchJQ(".attributes.state", "ready"))
				})
			})
		})
		It("Creates basic cluster", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.14.0"))
		})

		Context("Creates cluster with etcd encryption", func() {
			BeforeEach(func() {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
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
								  "oidc_endpoint_url": "https://127.0.0.1",
								  "thumbprint": "111111",
								  "role_arn": "",
								  "support_role_arn": "",
								  "instance_iam_roles" : {
									"worker_role_arn" : ""
								  },
								  "operator_role_prefix" : "test"
							  },
							  "etcd_encryption": {
								"kms_key_arn": "arn:aws:kms:us-west-2:111122223333:key/mrk-78dcc31c5865498cbe98ad5ab9769a04"
							  }
						  }
						},
						{
							"op": "add",
							"path": "/etcd_encryption",
							"value": true
						}]`),
					),
				)
			})
			When("Required together etcd_encryption and etcd_kms_key_arn", func() {
				It("fails when no etcd_kms_key_arn", func() {
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					etcd_encryption = true
				}`)
					Expect(terraform.Apply()).ToNot(BeZero())
				})
				It("fails when no etcd_encryption", func() {
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					etcd_kms_key_arn = "kms"
				}`)
					Expect(terraform.Apply()).ToNot(BeZero())
				})
			})
			It("Creates cluster with etcd encryption", func() {
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					etcd_encryption = true
					etcd_kms_key_arn = "arn:aws:kms:us-west-2:111122223333:key/mrk-78dcc31c5865498cbe98ad5ab9769a04"
				}`)
				Expect(terraform.Apply()).To(BeZero())
				resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
				Expect(resource).To(MatchJQ(".attributes.current_version", "4.14.0"))
			})
		})

		It("Creates basic cluster - and reconcile on a 404", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.14.0"))
			Expect(resource).To(MatchJQ(".attributes.id", "123")) // cluster has id 123

			// Prepare the server for reconcile
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.14.0"))
			Expect(resource).To(MatchJQ(".attributes.id", "1234")) // reconciled cluster has id of 1234
		})

		It("Creates basic cluster with properties", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),

					VerifyJQ(`.properties.`+propKey, propValue),
					RespondWithPatchedJSON(http.StatusCreated, template, fmt.Sprintf(`[
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
					},
                    {
                      "op": "add",
                      "path": "/properties",
                      "value": {
                        "%s": "%s"
                      }
                    }]`, propKey, propValue)),
				),
			)

			// Run the apply command:
			terraform.Source(fmt.Sprintf(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				properties = { 
					%s = "%s"
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`, propKey, propValue))
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates basic cluster with properties and update them", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),

					VerifyJQ(`.properties.`+propKey, propValue),
					RespondWithPatchedJSON(http.StatusCreated, template, fmt.Sprintf(`[
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
					},
                    {
                      "op": "add",
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"%s",
                        "rosa_tf_version":"%s",
                        "%s": "%s"
                      }
                    }]`, build.Commit, build.Version, propKey, propValue)),
				),
			)

			// Run the apply command:
			terraform.Source(fmt.Sprintf(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				properties = { 
					%s = "%s"
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`, propKey, propValue))
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+propKey, propValue))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+propKey, propValue))

			// Prepare server for update
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, fmt.Sprintf(`[
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
					},
					{
						"op": "add",
						"path": "/properties",
						"value": {
						  "rosa_tf_commit":"%s",
						  "rosa_tf_version":"%s",
						  "%s": "%s"
						}
					}]`, build.Commit, build.Version, propKey, propValue)),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, cluster123Route),
					VerifyJQ(`.properties.`+propKey, propValue+"_1"),
					RespondWithPatchedJSON(http.StatusOK, template, fmt.Sprintf(`[
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
					},
					{
						"op": "add",
						"path": "/properties",
						"value": {
						  "rosa_tf_commit":"%s",
						  "rosa_tf_version":"%s",
						  "%s": "%s"
						}
					}]`, build.Commit, build.Version, propKey, propValue+"_1")),
				),
			)

			// Run the apply command:
			terraform.Source(fmt.Sprintf(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				properties = {
					%s = "%s"
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`, propKey, propValue+"_1"))
			Expect(terraform.Apply()).To(BeZero())
			resource = terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+propKey, propValue+"_1"))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+propKey, propValue+"_1"))
		})

		It("Creates basic cluster with properties and delete them", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					VerifyJQ(`.name`, "my-cluster"),
					VerifyJQ(`.cloud_provider.id`, "aws"),
					VerifyJQ(`.region.id`, "us-west-1"),
					VerifyJQ(`.product.id`, "rosa"),
					VerifyJQ(`.properties.`+propKey, propValue),
					RespondWithPatchedJSON(http.StatusCreated, template, fmt.Sprintf(`[
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
					},
                    {
						"op": "add",
						"path": "/properties",
						"value": {
						  "rosa_tf_commit":"%s",
						  "rosa_tf_version":"%s",
						  "%s": "%s"
						}
					}]`, build.Commit, build.Version, propKey, propValue)),
				),
			)

			// Run the apply command:
			terraform.Source(fmt.Sprintf(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				properties = { 
					%s = "%s"
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`, propKey, propValue))
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties.`+propKey, propValue))
			Expect(resource).To(MatchJQ(`.attributes.properties.`+propKey, propValue))
			Expect(resource).To(MatchJQ(`.attributes.properties| keys | length`, 1))
			Expect(resource).To(MatchJQ(`.attributes.ocm_properties| keys | length`, 3))

			// Prepare server for update
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, fmt.Sprintf(`[
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
					},
					{
						"op": "add",
						"path": "/properties",
						"value": {
						  "rosa_tf_commit":"%s",
						  "rosa_tf_version":"%s",
						  "%s": "%s"
						}
					}]`, build.Commit, build.Version, propKey, propValue)),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPatch, cluster123Route),
					VerifyJQ(`.properties.`+propKey, nil),
					RespondWithPatchedJSON(http.StatusOK, template, fmt.Sprintf(`[
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
					},
                    {
						"op": "add",
						"path": "/properties",
						"value": {
						  "rosa_tf_commit":"%s",
						  "rosa_tf_version":"%s"
						}
					}]`, build.Commit, build.Version)),
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
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
					RespondWithJSON(http.StatusOK, versionListPage),
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})

		It("Should fail cluster creation when cluster name length is more than 54", func() {
			// Run the apply command:
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster-234567-foobar-foobar-foobar-foobar-fooobaaar-fooobaaz"
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).ToNot(BeZero())

		})

		Context("Test destroy cluster", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
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
						VerifyRequest(http.MethodGet, cluster123Route),
						RespondWithJSON(http.StatusOK, template),
					),
					CombineHandlers(
						VerifyRequest(http.MethodDelete, cluster123Route),
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
					 availability_zones = [
						 "us-west-1a",
						 "us-west-1b",
						 "us-west-1c",
					 ]
				}`)
				// it should return a warning so exit code will be "0":
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})

			It("Wait in destroy resource but use the default timeout", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
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
					  availability_zones = [
						  "us-west-1a",
						  "us-west-1b",
						  "us-west-1c",
					  ]
				}`)
				// it should return a warning so exit code will be "0":
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})

			It("Wait in destroy resource and set timeout to a negative value", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
				}`)
				// it should return a warning so exit code will be "0":
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})

			It("Wait in destroy resource and set timeout to a positive value", func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
				  }`)
				// it should return a warning so exit code will be "0":
				Expect(terraform.Apply()).To(BeZero())
				Expect(terraform.Destroy()).To(BeZero())
			})
		})

		Context("Test Proxy", func() {
			It("Creates cluster with http proxy and update it", func() {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
						RespondWithJSON(http.StatusOK, versionListPage),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
				}`)
				Expect(terraform.Apply()).To(BeZero())

				// apply for update the proxy's attributes
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
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
							"additional_trust_bundle" : "REDACTED"
						}
						}]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPatch, cluster123Route),
						VerifyJQ(`.proxy.https_proxy`, "https://proxy2.com"),
						VerifyJQ(`.proxy.no_proxy`, "test"),
						VerifyJQ(`.additional_trust_bundle`, "123"),
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
							"additional_trust_bundle" : "REDACTED"
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
				}`)
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
						RespondWithJSON(http.StatusOK, versionListPage),
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
				}`)
				Expect(terraform.Apply()).To(BeZero())

				// apply for update the proxy's attributes
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
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
					CombineHandlers(
						VerifyRequest(http.MethodPatch, cluster123Route),
						VerifyJQ(`.additional_trust_bundle`, "123"),
						RespondWithPatchedJSON(http.StatusCreated, templateWithTrustBundle, `[
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
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
				}`)
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
				proxy = {}
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			 }`)
			Expect(terraform.Apply()).NotTo(BeZero())

			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
								"oidc_endpoint_url": "https://127.0.0.1",
								"thumbprint": "111111",
								"role_arn": "",
								"support_role_arn": "",
								"instance_iam_roles": {
									"worker_role_arn": ""
								},
								"operator_role_prefix": "test"
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).To(BeZero())

		})
		It("Creates private cluster with aws subnet ids & private link", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
      						"us-west-1a",
							"us-west-1b",
							"us-west-1c"
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
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				private = true
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates cluster when private link is false", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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

			// Run the apply command:
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				private = false
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Creates rosa sts cluster with OIDC Configuration ID", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
		  	}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Fails to create cluster with incompatible account role's version and fail", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
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
							  "oidc_endpoint_url": "https://127.0.0.1",
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
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
			}`)
			// expect to get an error
			Expect(terraform.Apply()).ToNot(BeZero())
		})
	})

	Context("Upgrade", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
					RespondWithJSON(http.StatusOK, versionListPage),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
					RespondWithPatchedJSON(http.StatusCreated, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
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
					},
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.0"
			}`)
			Expect(terraform.Apply()).To(BeZero())
			// Verify initial cluster version
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.14.0"))
		})
		It("Upgrades Cluster", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
							}
						}
					]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
							}
						}
					]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.14.1"),
					RespondWithJSON(http.StatusOK, v4141Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies"),
					RespondWithJSON(http.StatusOK, emptyControlPlaneUpgradePolicies),
				),
				// Look for gate agreements by posting an upgrade policy w/ dryRun
				CombineHandlers(
					VerifyRequest(http.MethodPost, cluster123Route+"/control_plane/upgrade_policies", "dryRun=true"),
					VerifyJQ(".version", "4.14.1"),
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
							"version_raw_id_prefix": "4.14",
							"label": "api.openshift.com/gate-sts",
							"value": "4.14",
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
					VerifyRequest(http.MethodPost, cluster123Route+"/gate_agreements"),
					VerifyJQ(".version_gate.id", "999"),
					RespondWithJSON(http.StatusCreated, `{
						"kind": "VersionGateAgreement",
						"id": "888",
						"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
						"version_gate": {
						"kind": "VersionGate",
						"id": "999",
						"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
						"version_raw_id_prefix": "4.14",
						"label": "api.openshift.com/gate-sts",
						"value": "4.14",
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
					VerifyRequest(http.MethodPost, cluster123Route+"/control_plane/upgrade_policies"),
					VerifyJQ(".version", "4.14.1"),
					RespondWithJSON(http.StatusCreated, `
					{
						"id": "123",
						"schedule_type": "manual",
						"upgrade_type": "OSD",
						"version": "4.14.1",
						"next_run": "2023-06-09T20:59:00Z",
						"cluster_id": "123",
						"enable_minor_version_upgrades": true
					}`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, cluster123Route),
					RespondWithPatchedJSON(http.StatusCreated, template, `
					[
						{
						  "op": "add",
						  "path": "/properties",
						  "value": {
							"rosa_tf_commit": "",
							"rosa_tf_version": ""
						  }
						}
					]`),
				),
			)
			// Perform upgrade w/ auto-ack of sts-only gate agreements
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = ""
					support_role_arn = ""
					instance_iam_roles = {
						worker_role_arn = ""
					}
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.1"
			}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Does nothing if upgrade is in progress to a different version than the desired", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
							}
						}
					]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
							}
						}
					]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.14.1"),
					RespondWithJSON(http.StatusOK, v4141Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies"),
					RespondWithJSON(http.StatusOK, `{
						"page": 1,
						"size": 1,
						"total": 1,
						"items": [
							{
								"id": "456",
								"schedule_type": "manual",
								"upgrade_type": "ControlPlane",
								"version": "4.14.0",
								"next_run": "2023-06-09T20:59:00Z",
								"cluster_id": "123",
								"enable_minor_version_upgrades": true
							}
						]
					}`),
				),
				// Check it's state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, `
					{
						"id": "456",
						"state": {
							"description": "Upgrade in progress",
							"value": "started"
						}
					}`),
				),
			)
			// Perform try the upgrade
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = ""
					support_role_arn = ""
					instance_iam_roles = {
						worker_role_arn = ""
					}
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.1"
			}`)
			// Will fail due to upgrade in progress
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		It("Cancels and upgrade for the wrong version & schedules new", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
							}
						}
					]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template, `
					[
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
						},
						{
							"op": "add",
							"path": "/properties",
							"value": {
								"rosa_tf_commit": "",
								"rosa_tf_version": ""
							}
						}
					]`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.14.1"),
					RespondWithJSON(http.StatusOK, v4141Info),
				),
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "UpgradePolicyState",
						"page": 1,
						"size": 0,
						"total": 0,
						"items": [
							{
								"id": "456",
								"schedule_type": "manual",
								"upgrade_type": "ControlPlane",
								"version": "4.14.0",
								"next_run": "2023-06-09T20:59:00Z",
								"cluster_id": "123",
								"enable_minor_version_upgrades": true
							}
						]
					}`),
				),
				// Check it's state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, `{
						"id": "456",
						"state": {
							"description": "",
							"value": "scheduled"
						}
					}`),
				),
				// Delete existing upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodDelete, cluster123Route+"/control_plane/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				// Look for gate agreements by posting an upgrade policy w/ dryRun (no gates necessary)
				CombineHandlers(
					VerifyRequest(http.MethodPost, cluster123Route+"/control_plane/upgrade_policies", "dryRun=true"),
					VerifyJQ(".version", "4.14.1"),
					RespondWithJSON(http.StatusNoContent, ""),
				),
				// Create an upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodPost, cluster123Route+"/control_plane/upgrade_policies"),
					VerifyJQ(".version", "4.14.1"),
					RespondWithJSON(http.StatusCreated, `{
						"id": "123",
						"schedule_type": "manual",
						"upgrade_type": "ControlPlane",
						"version": "4.14.1",
						"next_run": "2023-06-09T20:59:00Z",
						"cluster_id": "123",
						"enable_minor_version_upgrades": true
					}`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, cluster123Route),
					RespondWithJSON(http.StatusOK, template),
				),
			)
			// Perform try the upgrade
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = ""
					support_role_arn = ""
					instance_iam_roles = {
						worker_role_arn = ""
					}
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.1"
			}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Cancels upgrade if version=current_version", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithJSON(http.StatusOK, template),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies"),
					RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"page": 1,
					"size": 0,
					"total": 0,
					"items": [
						{
							"id": "456",
							"schedule_type": "manual",
							"upgrade_type": "ControlPlane",
							"version": "4.14.1",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
						}
					]
				}`),
				),
				// Check it's state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, `{
						"id": "456",
						"state": {
							"description": "",
							"value": "scheduled"
						}
					}`),
				),
				// Delete existing upgrade policy
				CombineHandlers(
					VerifyRequest(http.MethodDelete, cluster123Route+"/control_plane/upgrade_policies/456"),
					RespondWithJSON(http.StatusOK, "{}"),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, cluster123Route),
					RespondWithJSON(http.StatusOK, template),
				),
			)
			// Set version to match current cluster version
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = ""
					support_role_arn = ""
					instance_iam_roles = {
						worker_role_arn = ""
					}
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.0"
			}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("is an error to request a version older than current", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template,
						`[
					{
						"op": "replace",
						"path": "/version/id",
						"value": "openshift-v4.14.2"
					}]`),
				),
			)
			// Set version to before current cluster version, but after version from create
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = ""
					support_role_arn = ""
					instance_iam_roles = {
						worker_role_arn = ""
					}
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.1"
			}`)
			Expect(terraform.Apply()).NotTo(BeZero())
		})

		It("older than current is allowed as long as not changed", func() {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
					RespondWithPatchedJSON(http.StatusOK, template,
						`[
						{
							"op": "replace",
							"path": "/version/id",
							"value": "openshift-v4.14.1"
						}]`),
				),
				// Patch the cluster (w/ no changes)
				CombineHandlers(
					VerifyRequest(http.MethodPatch, cluster123Route),
					RespondWithJSON(http.StatusOK, template),
				),
			)
			// Set version to before current cluster version, but matching what was
			// used during creation (i.e. in state file)
			terraform.Source(`
			resource "rhcs_cluster_rosa_hcp" "my_cluster" {
				name           = "my-cluster"
				cloud_region   = "us-west-1"
				aws_account_id = "123"
				aws_billing_account_id = "123"
				sts = {
					operator_role_prefix = "test"
					role_arn = ""
					support_role_arn = ""
					instance_iam_roles = {
						worker_role_arn = ""
					}
				}
				aws_subnet_ids = [
					"id1", "id2", "id3"
				]
				availability_zones = [
					"us-west-1a",
					"us-west-1b",
					"us-west-1c",
				]
				version = "4.14.0"
			}`)
			Expect(terraform.Apply()).To(BeZero())
		})

		Context("Un-acked gates", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					// Refresh cluster state
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
						RespondWithPatchedJSON(http.StatusOK, template, `
						[
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
							},
							{
								"op": "add",
								"path": "/properties",
								"value": {
									"rosa_tf_commit": "",
									"rosa_tf_version": ""
								}
							}
						]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route),
						RespondWithPatchedJSON(http.StatusOK, template, `
						[
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
							},
							{
								"op": "add",
								"path": "/properties",
								"value": {
									"rosa_tf_commit": "",
									"rosa_tf_version": ""
								}
							}
						]`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.14.1"),
						RespondWithJSON(http.StatusOK, v4141Info),
					),
					// Look for existing upgrade policies
					CombineHandlers(
						VerifyRequest(http.MethodGet, cluster123Route+"/control_plane/upgrade_policies"),
						RespondWithJSON(http.StatusOK, emptyControlPlaneUpgradePolicies),
					),
					// Look for gate agreements by posting an upgrade policy w/ dryRun
					CombineHandlers(
						VerifyRequest(http.MethodPost, cluster123Route+"/control_plane/upgrade_policies", "dryRun=true"),
						VerifyJQ(".version", "4.14.1"),
						RespondWithJSON(http.StatusBadRequest, `{
							"kind": "Error",
							"id": "400",
							"href": "/api/clusters_mgmt/v1/errors/400",
							"code": "CLUSTERS-MGMT-400",
							"reason": "There are missing version gate agreements for this cluster. See details.",
							"details": [
							{
								"id": "999",
								"version_raw_id_prefix": "4.14",
								"label": "api.openshift.com/ackme",
								"value": "4.14",
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
				resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					sts = {
						operator_role_prefix = "test"
						role_arn = ""
						support_role_arn = ""
						instance_iam_roles = {
							worker_role_arn = ""
						}
					}
					aws_subnet_ids = [
						"id1", "id2", "id3"
					]
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					version = "4.14.1"
				}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
			It("Fails upgrade if wrong version is acked", func() {
				terraform.Source(`
				resource "rhcs_cluster_rosa_hcp" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					aws_billing_account_id = "123"
					sts = {
						operator_role_prefix = "test"
						role_arn = ""
						support_role_arn = ""
						instance_iam_roles = {
							worker_role_arn = ""
						}
					}
					aws_subnet_ids = [
						"id1", "id2", "id3"
					]
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					version = "4.14.1"
					upgrade_acknowledgements_for = "1.1"
				}`)
				Expect(terraform.Apply()).NotTo(BeZero())
			})
			It("It acks gates if correct ack is provided", func() {
				server.AppendHandlers(
					// Send acks for all gate agreements
					CombineHandlers(
						VerifyRequest(http.MethodPost, cluster123Route+"/gate_agreements"),
						VerifyJQ(".version_gate.id", "999"),
						RespondWithJSON(http.StatusCreated, `{
						"kind": "VersionGateAgreement",
						"id": "888",
						"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
						"version_gate": {
						"id": "999",
						"version_raw_id_prefix": "4.14",
						"label": "api.openshift.com/gate-sts",
						"value": "4.14",
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
						VerifyRequest(http.MethodPost, cluster123Route+"/control_plane/upgrade_policies"),
						VerifyJQ(".version", "4.14.1"),
						RespondWithJSON(http.StatusCreated, `
						{
							"kind": "UpgradePolicy",
							"id": "123",
							"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
							"schedule_type": "manual",
							"upgrade_type": "ControlPlane",
							"version": "4.14.1",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
						}`),
					),
					// Patch the cluster (w/ no changes)
					CombineHandlers(
						VerifyRequest(http.MethodPatch, cluster123Route),
						RespondWithJSON(http.StatusCreated, template),
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
						role_arn = ""
						support_role_arn = ""
						instance_iam_roles = {
							worker_role_arn = ""
						}
					}
					aws_subnet_ids = [
						"id1", "id2", "id3"
					]
					availability_zones = [
						"us-west-1a",
						"us-west-1b",
						"us-west-1c",
					]
					version = "4.14.1"
					upgrade_acknowledgements_for = "4.14"
				}`)
				Expect(terraform.Apply()).To(BeZero())
			})
		})
	})

	Context("Import", func() {
		It("can import a cluster", func() {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, cluster123Route),
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
			)

			// Run the apply command:
			terraform.Source(`resource "rhcs_cluster_rosa_hcp" "my_cluster" {}`)
			Expect(terraform.Import("rhcs_cluster_rosa_hcp.my_cluster", "123")).To(BeZero())
			resource := terraform.Resource("rhcs_cluster_rosa_hcp", "my_cluster")
			Expect(resource).To(MatchJQ(".attributes.current_version", "4.14.0"))
		})
	})
})
