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

package hyperfleet

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/openshift-online/ocm-sdk-go/testing"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_nodepool_hyperfleet", func() {
	Context("Create operation", func() {
		var hyperfleetServer *Server

		BeforeEach(func() {
			hyperfleetServer = NewServer()
		})

		AfterEach(func() {
			hyperfleetServer.Close()
		})

		It("creates a nodepool successfully", func() {
			clusterResponse := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "test-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"rolesRef": {
									"nodePoolManagementARN": "arn:aws:iam::123456789012:role/test-cluster-NodePool"
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.test-cluster.example.com",
						"port": 6443
					}
				}
			}`

			nodepoolResponse := `{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": true,
					"ec2InstanceMetadataHttpTokens": "required",
					"labels": {
						"workload-type": "general"
					},
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge",
								"rootVolume": {
									"size": 100
								}
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Provisioning"
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					RespondWith(http.StatusCreated, nodepoolResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, nodepoolResponse, header),
				),
			)

			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider         = rhcs.hf
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
					labels = {
						"workload-type" = "general"
					}
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("Upgrade disruption budget", func() {
		var hyperfleetServer *Server

		// The parent cluster PostExpand reads to derive the worker instance
		// profile.
		clusterResponse := `{
			"metadata": {
				"uid": "test-cluster-id",
				"name": "test-cluster",
				"creationTimestamp": "2024-01-01T00:00:00Z"
			},
			"spec": {
				"hostedCluster": {
					"platform": {
						"type": "AWS",
						"aws": {
							"region": "us-east-1",
							"rolesRef": {
								"nodePoolManagementARN": "arn:aws:iam::123456789012:role/test-cluster-NodePool"
							}
						}
					}
				}
			},
			"status": {
				"phase": "Ready",
				"controlPlaneEndpoint": {
					"host": "api.test-cluster.example.com",
					"port": 6443
				}
			}
		}`

		// nodepoolResponse is the Platform API representation of the node pool
		// with the given spec.maxUnavailable (raw JSON, may be empty).
		nodepoolResponse := func(maxUnavailable string) string {
			return fmt.Sprintf(`{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": true,
					%s
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge"
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Ready"
				}
			}`, maxUnavailable)
		}

		header := http.Header{"Content-Type": []string{"application/json"}}

		// nodepoolSource renders the node pool configuration with the given
		// max_unavailable value (HCL, may be empty).
		nodepoolSource := func(url, maxUnavailable string) string {
			return fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider      = rhcs.hf
					cluster_id    = "test-cluster-id"
					cluster_name  = "test-cluster"
					name          = "test-cluster.worker"
					replicas      = 3
					subnet_id     = "subnet-0abc123"
					auto_repair   = true
					instance_type = "m5.xlarge"
					%s
				}
			`, url, maxUnavailable)
		}

		BeforeEach(func() {
			hyperfleetServer = NewServer()
		})

		AfterEach(func() {
			hyperfleetServer.Close()
		})

		It("sends an absolute node count as spec.maxUnavailable on create", func() {
			response := nodepoolResponse(`"maxUnavailable": 2,`)
			hyperfleetServer.RouteToHandler(
				http.MethodGet, "/api/v0/clusters/test-cluster",
				RespondWith(http.StatusOK, clusterResponse, header),
			)
			hyperfleetServer.RouteToHandler(
				http.MethodGet, "/api/v0/nodepools/test-cluster.worker",
				RespondWith(http.StatusOK, response, header),
			)
			hyperfleetServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					VerifyJQ(`.spec.maxUnavailable`, 2.0),
					RespondWith(http.StatusCreated, response, header),
				),
			)

			Terraform.Source(nodepoolSource(hyperfleetServer.URL(), `max_unavailable = "2"`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// The value is an int-or-string on the API side, so a second apply
			// must be a no-op: a requoted or renumbered value coming back would
			// plan a diff.
			runOutput = Terraform.Run("plan", "-no-color", "-detailed-exitcode")
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates spec.maxUnavailable in place", func() {
			created := nodepoolResponse(`"maxUnavailable": 1,`)
			updated := nodepoolResponse(`"maxUnavailable": 2,`)
			hyperfleetServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					RespondWith(http.StatusCreated, created, header),
				),
				// Second apply: refresh (Read), then update. PostExpand fetches
				// the parent cluster again before Update.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, created, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, created, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/nodepools/test-cluster.worker"),
					VerifyJQ(`.spec.maxUnavailable`, 2.0),
					RespondWith(http.StatusOK, updated, header),
				),
			)

			Terraform.Source(nodepoolSource(hyperfleetServer.URL(), `max_unavailable = "1"`))
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			Terraform.Source(nodepoolSource(hyperfleetServer.URL(), `max_unavailable = "2"`))
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("rejects a value that is not a node count", func() {
			Terraform.Source(nodepoolSource(hyperfleetServer.URL(), `max_unavailable = "two"`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("must be a non-negative whole number of nodes")
		})

		// The Platform API accepts a percentage, but pathbind cannot round-trip
		// one into the SDK's intstr value yet, so the provider rejects it up
		// front with an explanation instead of failing inside Expand. Replace
		// this with round-trip coverage once the clientset handles intstr.
		It("rejects a percentage with an explanation", func() {
			Terraform.Source(nodepoolSource(hyperfleetServer.URL(), `max_unavailable = "30%"`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("percentages are not supported yet")
		})

		It("rejects a node count above the replica count", func() {
			Terraform.Source(nodepoolSource(hyperfleetServer.URL(), `max_unavailable = "4"`))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("max_unavailable 4 cannot exceed replicas 3")
		})
	})

	Context("Update operation", func() {
		var hyperfleetServer *Server

		BeforeEach(func() {
			hyperfleetServer = NewServer()
		})

		AfterEach(func() {
			hyperfleetServer.Close()
		})

		It("updates nodepool labels", func() {
			clusterResponse := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "test-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"rolesRef": {
									"nodePoolManagementARN": "arn:aws:iam::123456789012:role/test-cluster-NodePool"
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.test-cluster.example.com",
						"port": 6443
					}
				}
			}`

			createResponse := `{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": true,
					"labels": {
						"workload-type": "general"
					},
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge",
								"rootVolume": {
									"size": 100
								}
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Ready"
				}
			}`

			updateResponse := `{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": true,
					"labels": {
						"workload-type": "general",
						"environment": "production"
					},
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge",
								"rootVolume": {
									"size": 100
								}
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Ready"
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				// First apply: create without extra labels
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					RespondWith(http.StatusCreated, createResponse, header),
				),
				// Second apply: refresh (Read), then update to add labels.
				// PostExpand fetches the parent cluster again before Update.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, updateResponse, header),
				),
			)

			// First apply: create with single label
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider         = rhcs.hf
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
					labels = {
						"workload-type" = "general"
					}
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Second apply: update to add environment label
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider         = rhcs.hf
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
					labels = {
						"workload-type" = "general",
						"environment"   = "production"
					}
				}
			`, hyperfleetServer.URL()))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates nodepool auto_repair", func() {
			clusterResponse := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "test-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"rolesRef": {
									"nodePoolManagementARN": "arn:aws:iam::123456789012:role/test-cluster-NodePool"
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.test-cluster.example.com",
						"port": 6443
					}
				}
			}`

			createResponse := `{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": true,
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge",
								"rootVolume": {
									"size": 100
								}
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Ready"
				}
			}`

			updateResponse := `{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": false,
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge",
								"rootVolume": {
									"size": 100
								}
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Ready"
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				// First apply: create with auto_repair=true
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					RespondWith(http.StatusCreated, createResponse, header),
				),
				// Second apply: refresh (Read), then update to disable auto_repair.
				// PostExpand fetches the parent cluster again before Update.
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, updateResponse, header),
				),
			)

			// First apply: create with auto_repair=true (default)
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider         = rhcs.hf
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Second apply: update to disable auto_repair
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider         = rhcs.hf
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = false
					instance_type    = "m5.xlarge"
					size = 100
				}
			`, hyperfleetServer.URL()))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("Delete operation", func() {
		var hyperfleetServer *Server

		BeforeEach(func() {
			hyperfleetServer = NewServer()
		})

		AfterEach(func() {
			hyperfleetServer.Close()
		})

		It("deletes a nodepool successfully", func() {
			clusterResponse := `{
				"metadata": {
					"uid": "test-cluster-id",
					"name": "test-cluster",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1",
								"rolesRef": {
									"nodePoolManagementARN": "arn:aws:iam::123456789012:role/test-cluster-NodePool"
								}
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.test-cluster.example.com",
						"port": 6443
					}
				}
			}`

			createResponse := `{
				"metadata": {
					"uid": "test-nodepool-id",
					"name": "test-cluster.worker",
					"creationTimestamp": "2024-01-01T00:00:00Z"
				},
				"spec": {
					"autoRepair": true,
					"nodePool": {
						"clusterName": "test-cluster",
						"release": {
							"image": ""
						},
						"platform": {
							"type": "AWS",
							"aws": {
								"subnet": {
									"id": "subnet-0abc123"
								},
								"instanceType": "m5.xlarge",
								"rootVolume": {
									"size": 100
								}
							}
						},
						"replicas": 3
					}
				},
				"status": {
					"phase": "Ready"
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
					RespondWith(http.StatusOK, clusterResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					RespondWith(http.StatusCreated, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusNoContent, ""),
				),
			)

			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_hyperfleet" "test" {
					provider         = rhcs.hf
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})
})
