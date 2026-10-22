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

package seaplane

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_nodepool_seaplane", func() {
	Context("Create operation", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
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

			seaplaneServer.AppendHandlers(
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
					alias          = "sp"
							seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_seaplane" "test" {
					provider         = rhcs.sp
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
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("Update operation", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
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

			seaplaneServer.AppendHandlers(
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
					alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_seaplane" "test" {
					provider         = rhcs.sp
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
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Second apply: update to add environment label
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_seaplane" "test" {
					provider         = rhcs.sp
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
			`, seaplaneServer.URL()))

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

			seaplaneServer.AppendHandlers(
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
					alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_seaplane" "test" {
					provider         = rhcs.sp
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Second apply: update to disable auto_repair
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_seaplane" "test" {
					provider         = rhcs.sp
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = false
					instance_type    = "m5.xlarge"
					size = 100
				}
			`, seaplaneServer.URL()))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("Delete operation", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
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

			seaplaneServer.AppendHandlers(
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
					alias          = "sp"
					seaplane_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_nodepool_seaplane" "test" {
					provider         = rhcs.sp
					cluster_id       = "test-cluster-id"
					cluster_name     = "test-cluster"
					name             = "test-cluster.worker"
					replicas         = 3
					subnet_id        = "subnet-0abc123"
					auto_repair      = true
					instance_type    = "m5.xlarge"
					size = 100
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})
})

		Context("Upgrade disruption budget", func() {
			var seaplaneServer *Server

			BeforeEach(func() {
				seaplaneServer = NewServer()
			})

			AfterEach(func() {
				seaplaneServer.Close()
			})

			It("sets max_unavailable as an absolute count on create", func() {
				clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"prefix": "roles-prefix"}}}}}}`
				createResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"maxUnavailable": "2", "autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`

				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.workers", CombineHandlers(
					RespondWith(http.StatusOK, createResponse, http.Header{"Content-Type": []string{"application/json"}}),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"),
						RespondWith(http.StatusOK, clusterResponse, http.Header{"Content-Type": []string{"application/json"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
						VerifyJSONRepresenting(map[string]interface{}{
							"metadata": map[string]interface{}{"name": "test-cluster.workers"},
							"spec": map[string]interface{}{
								"maxUnavailable": "2",
								"autoRepair":     true,
								"nodePool": map[string]interface{}{
									"clusterName": "test-cluster",
									"replicas":    float64(5),
									"platform": map[string]interface{}{
										"aws": map[string]interface{}{
											"instanceType": "m5.xlarge",
											"subnet":       map[string]interface{}{"id": "subnet-0abc123"},
										},
									},
								},
							},
						}),
						RespondWith(http.StatusCreated, createResponse, http.Header{"Content-Type": []string{"application/json"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.workers"),
						RespondWith(http.StatusNoContent, ""),
					),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
					resource "rhcs_nodepool_seaplane" "test" {
						cluster_id       = "c-123"
						cluster_name     = "test-cluster"
						name             = "test-cluster.workers"
						replicas         = 5
						max_unavailable  = "2"
						auto_repair      = true
						subnet_id        = "subnet-0abc123"
						instance_type    = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				planOutput := Terraform.Run("plan", "-detailed-exitcode")
				Expect(planOutput.ExitCode).To(BeZero(), "expected no-op plan after create")

				runOutput = Terraform.Destroy()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("updates max_unavailable in place", func() {
				clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"prefix": "roles-prefix"}}}}}}`
				createResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"maxUnavailable": "1", "autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
				updateResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"maxUnavailable": "2", "autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`

				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.workers", CombineHandlers(
					RespondWith(http.StatusOK, createResponse, http.Header{"Content-Type": []string{"application/json"}}),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse)),
					CombineHandlers(VerifyRequest(http.MethodPost, "/api/v0/nodepools"), RespondWith(http.StatusCreated, createResponse)),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
					resource "rhcs_nodepool_seaplane" "test" {
						cluster_id       = "c-123"
						cluster_name     = "test-cluster"
						name             = "test-cluster.workers"
						replicas         = 5
						max_unavailable  = "1"
						auto_repair      = true
						subnet_id        = "subnet-0abc123"
						instance_type    = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.workers", CombineHandlers(
					RespondWith(http.StatusOK, updateResponse, http.Header{"Content-Type": []string{"application/json"}}),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/api/v0/nodepools/test-cluster.workers"),
						VerifyJSONRepresenting(map[string]interface{}{
							"metadata": map[string]interface{}{"uid": "np-456", "name": "test-cluster.workers"},
							"spec": map[string]interface{}{
								"maxUnavailable": "2",
								"autoRepair":     true,
								"nodePool": map[string]interface{}{
									"clusterName": "test-cluster",
									"replicas":    float64(5),
									"platform": map[string]interface{}{
										"aws": map[string]interface{}{
											"instanceType":    "m5.xlarge",
											"instanceProfile": "roles-prefix-ROSA-Worker-Role",
											"subnet":          map[string]interface{}{"id": "subnet-0abc123"},
										},
									},
								},
							},
						}),
						RespondWith(http.StatusOK, updateResponse, http.Header{"Content-Type": []string{"application/json"}}),
					),
					CombineHandlers(VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.workers"), RespondWith(http.StatusNoContent, "")),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
					resource "rhcs_nodepool_seaplane" "test" {
						cluster_id       = "c-123"
						cluster_name     = "test-cluster"
						name             = "test-cluster.workers"
						replicas         = 5
						max_unavailable  = "2"
						auto_repair      = true
						subnet_id        = "subnet-0abc123"
						instance_type    = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput = Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				runOutput = Terraform.Destroy()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("rejects non-integer max_unavailable", func() {
				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
					resource "rhcs_nodepool_seaplane" "test" {
						cluster_id       = "c-123"
						cluster_name     = "test-cluster"
						name             = "test-cluster.workers"
						replicas         = 5
						max_unavailable  = "abc"
						auto_repair      = true
						subnet_id        = "subnet-0abc123"
						instance_type    = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("non-negative whole number")
			})

			It("rejects percentage max_unavailable with explanation", func() {
				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
					resource "rhcs_nodepool_seaplane" "test" {
						cluster_id       = "c-123"
						cluster_name     = "test-cluster"
						name             = "test-cluster.workers"
						replicas         = 5
						max_unavailable  = "50%%"
						auto_repair      = true
						subnet_id        = "subnet-0abc123"
						instance_type    = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("percentages are not supported yet")
			})

			It("rejects max_unavailable exceeding replicas", func() {
				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
					resource "rhcs_nodepool_seaplane" "test" {
						cluster_id       = "c-123"
						cluster_name     = "test-cluster"
						name             = "test-cluster.workers"
						replicas         = 5
						max_unavailable  = "10"
						auto_repair      = true
						subnet_id        = "subnet-0abc123"
						instance_type    = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("cannot exceed replicas")
			})
		})
	})
})
