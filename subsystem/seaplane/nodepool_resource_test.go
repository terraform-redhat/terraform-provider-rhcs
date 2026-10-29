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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/ginkgo/v2/dsl/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"

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
					func(_ http.ResponseWriter, req *http.Request) {
						var body v1alpha1.NodePool
						Expect(json.NewDecoder(req.Body).Decode(&body)).To(Succeed())
						Expect(body.Spec.NodePool.Platform.AWS).NotTo(BeNil())
						Expect(body.Spec.NodePool.Platform.AWS.RootVolume).NotTo(BeNil())
						Expect(body.Spec.NodePool.Platform.AWS.RootVolume.Size).To(Equal(int64(100)))
					},
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
					disk_size = 100
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
					func(_ http.ResponseWriter, req *http.Request) {
						var body v1alpha1.NodePool
						Expect(json.NewDecoder(req.Body).Decode(&body)).To(Succeed())
						Expect(body.Spec.NodePool.Platform.AWS).NotTo(BeNil())
						Expect(body.Spec.NodePool.Platform.AWS.RootVolume).NotTo(BeNil())
						Expect(body.Spec.NodePool.Platform.AWS.RootVolume.Size).To(Equal(int64(100)))
					},
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
					disk_size = 100
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
					disk_size = 100
					labels = {
						"workload-type" = "general",
						"environment"   = "production"
					}
				}
			`, seaplaneServer.URL()))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates nodepool root tags in place", func() {
			clusterResponse := `{
				"metadata": {"uid":"test-cluster-id","name":"test-cluster","creationTimestamp":"2024-01-01T00:00:00Z"},
				"spec": {"hostedCluster":{"platform":{"type":"AWS","aws":{"region":"us-east-1","rolesRef":{"nodePoolManagementARN":"arn:aws:iam::123456789012:role/test-cluster-NodePool"}}}}},
				"status": {"phase":"Ready","controlPlaneEndpoint":{"host":"api.test-cluster.example.com","port":6443}}
			}`
			createResponse := `{
				"metadata": {"uid":"test-nodepool-id","name":"test-cluster.worker","creationTimestamp":"2024-01-01T00:00:00Z"},
				"spec": {
					"autoRepair":true,
					"tags":{"team":"before"},
					"nodePool":{"clusterName":"test-cluster","release":{"image":""},"platform":{"type":"AWS","aws":{"subnet":{"id":"subnet-0abc123"},"instanceType":"m5.xlarge","rootVolume":{"size":100}}},"replicas":3}
				},
				"status":{"phase":"Ready"}
			}`
			updateResponse := strings.Replace(createResponse, `"before"`, `"after"`, 1)
			header := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.RouteToHandler(
				http.MethodGet, "/api/v0/clusters/test-cluster",
				RespondWith(http.StatusOK, clusterResponse, header),
			)
			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					verifyJQ(`.spec.tags.team`, "before"),
					RespondWith(http.StatusCreated, createResponse, header),
				),
			)

			resourceConfig := func(tagValue string) string {
				return fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123456789012"
						aws_region     = "us-east-1"
					}

					resource "rhcs_nodepool_seaplane" "test" {
						provider      = rhcs.sp
						cluster_id    = "test-cluster-id"
						cluster_name  = "test-cluster"
						name          = "test-cluster.worker"
						replicas      = 3
						subnet_id     = "subnet-0abc123"
						auto_repair   = true
						instance_type = "m5.xlarge"
						disk_size          = 100
						tags = { team = %q }
					}
				`, seaplaneServer.URL(), tagValue)
			}

			Terraform.Source(resourceConfig("before"))
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			seaplaneServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/nodepools/test-cluster.worker"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/nodepools/test-cluster.worker"),
					verifyJQ(`.spec.tags.team`, "after"),
					RespondWith(http.StatusOK, updateResponse, header),
				),
			)

			Terraform.Source(resourceConfig("after"))
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
					disk_size = 100
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
					disk_size = 100
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
					disk_size = 100
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("Windows License Included image type", func() {
		var seaplaneServer *Server

		BeforeEach(func() {
			seaplaneServer = NewServer()
		})

		AfterEach(func() {
			seaplaneServer.Close()
		})

		It("sends imageType: Windows and produces clean re-plan", func() {
			clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"nodePoolManagementARN": "arn:aws:iam::123456789012:role/roles-prefix-node-pool-management"}}}}}}`
			createResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.windows-pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "platform": {"type": "AWS", "aws": {"imageType": "Windows", "instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
			jsonHeader := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.windows-pool", CombineHandlers(
				RespondWith(http.StatusOK, createResponse, jsonHeader),
			))
			seaplaneServer.AppendHandlers(
				CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					verifyJQ(`.spec.nodePool.platform.aws.imageType`, "Windows"),
					verifyJQ(`.spec.nodePool.platform.aws.instanceProfile`, "roles-prefix-ROSA-Worker-Role"),
					RespondWith(http.StatusCreated, createResponse, jsonHeader),
				),
				CombineHandlers(VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.windows-pool"), RespondWith(http.StatusNoContent, "")),
			)

			Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider      = rhcs.sp
						cluster_id    = "c-123"
						cluster_name  = "test-cluster"
						name          = "test-cluster.windows-pool"
						replicas      = 2
						image_type    = "Windows"
						auto_repair   = true
						subnet_id     = "subnet-0abc123"
						instance_type = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			planOutput := Terraform.Run("plan", "-detailed-exitcode")
			Expect(planOutput.ExitCode).To(BeZero(), "expected no-op plan after create")

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("replaces the pool when image_type changes", func() {
			clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"nodePoolManagementARN": "arn:aws:iam::123456789012:role/roles-prefix-node-pool-management"}}}}}}`
			linuxResponse := `{"metadata": {"uid": "np-linux", "name": "test-cluster.pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "platform": {"type": "AWS", "aws": {"imageType": "Linux", "instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
			windowsResponse := `{"metadata": {"uid": "np-windows", "name": "test-cluster.pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "platform": {"type": "AWS", "aws": {"imageType": "Windows", "instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
			jsonHeader := http.Header{"Content-Type": []string{"application/json"}}

			seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.pool", CombineHandlers(
				RespondWith(http.StatusOK, linuxResponse, jsonHeader),
			))
			seaplaneServer.AppendHandlers(
				CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					verifyJQ(`.spec.nodePool.platform.aws.imageType`, "Linux"),
					RespondWith(http.StatusCreated, linuxResponse, jsonHeader),
				),
			)

			Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider      = rhcs.sp
						cluster_id    = "c-123"
						cluster_name  = "test-cluster"
						name          = "test-cluster.pool"
						replicas      = 2
						image_type    = "Linux"
						auto_repair   = true
						subnet_id     = "subnet-0abc123"
						instance_type = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.pool", CombineHandlers(
				RespondWith(http.StatusOK, windowsResponse, jsonHeader),
			))
			seaplaneServer.AppendHandlers(
				CombineHandlers(VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.pool"), RespondWith(http.StatusNoContent, "")),
				CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
					verifyJQ(`.spec.nodePool.platform.aws.imageType`, "Windows"),
					RespondWith(http.StatusCreated, windowsResponse, jsonHeader),
				),
				CombineHandlers(VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.pool"), RespondWith(http.StatusNoContent, "")),
			)

			Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider      = rhcs.sp
						cluster_id    = "c-123"
						cluster_name  = "test-cluster"
						name          = "test-cluster.pool"
						replicas      = 2
						image_type    = "Windows"
						auto_repair   = true
						subnet_id     = "subnet-0abc123"
						instance_type = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("rejects bad enum value", func() {
			Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider      = rhcs.sp
						cluster_id    = "c-123"
						cluster_name  = "test-cluster"
						name          = "test-cluster.pool"
						replicas      = 2
						image_type    = "windows"
						auto_repair   = true
						subnet_id     = "subnet-0abc123"
						instance_type = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`must be "Linux" or "Windows"`)
		})

		Context("node_drain_grace_period validation", func() {
			It("accepts valid duration strings", func() {
				clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"nodePoolManagementARN": "arn:aws:iam::123456789012:role/roles-prefix-node-pool-management"}}}}}}`
				nodepoolResponse := `{"metadata": {"uid": "np-123", "name": "test-cluster.pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "nodeDrainTimeout": "5m", "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
				jsonHeader := http.Header{"Content-Type": []string{"application/json"}}

				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.pool", CombineHandlers(
					RespondWith(http.StatusOK, nodepoolResponse, jsonHeader),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
						verifyJQ(`.spec.nodePool.nodeDrainTimeout`, "5m"),
						RespondWith(http.StatusCreated, nodepoolResponse, jsonHeader),
					),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider           = rhcs.sp
						cluster_id         = "c-123"
						cluster_name       = "test-cluster"
						name               = "test-cluster.pool"
						replicas           = 2
						node_drain_grace_period ="5m"
						auto_repair        = true
						subnet_id          = "subnet-0abc123"
						instance_type      = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				planOutput := Terraform.Run("plan", "-detailed-exitcode")
				Expect(planOutput.ExitCode).To(BeZero(), "expected no-op plan after create")

				runOutput = Terraform.Destroy()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("accepts various duration formats", func() {
				clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"nodePoolManagementARN": "arn:aws:iam::123456789012:role/roles-prefix-node-pool-management"}}}}}}`
				nodepoolResponse := `{"metadata": {"uid": "np-123", "name": "test-cluster.pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "nodeDrainTimeout": "1h30m", "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
				jsonHeader := http.Header{"Content-Type": []string{"application/json"}}

				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.pool", CombineHandlers(
					RespondWith(http.StatusOK, nodepoolResponse, jsonHeader),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
						verifyJQ(`.spec.nodePool.nodeDrainTimeout`, "1h30m"),
						RespondWith(http.StatusCreated, nodepoolResponse, jsonHeader),
					),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider           = rhcs.sp
						cluster_id         = "c-123"
						cluster_name       = "test-cluster"
						name               = "test-cluster.pool"
						replicas           = 2
						node_drain_grace_period ="1h30m"
						auto_repair        = true
						subnet_id          = "subnet-0abc123"
						instance_type      = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				runOutput = Terraform.Destroy()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("allows in-place update of node_drain_grace_period", func() {
				clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"nodePoolManagementARN": "arn:aws:iam::123456789012:role/roles-prefix-node-pool-management"}}}}}}`
				responseWith5m := `{"metadata": {"uid": "np-123", "name": "test-cluster.pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "nodeDrainTimeout": "5m", "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
				responseWith10m := `{"metadata": {"uid": "np-123", "name": "test-cluster.pool"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 2, "nodeDrainTimeout": "10m", "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
				jsonHeader := http.Header{"Content-Type": []string{"application/json"}}

				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.pool", CombineHandlers(
					RespondWith(http.StatusOK, responseWith5m, jsonHeader),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
						verifyJQ(`.spec.nodePool.nodeDrainTimeout`, "5m"),
						RespondWith(http.StatusCreated, responseWith5m, jsonHeader),
					),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider           = rhcs.sp
						cluster_id         = "c-123"
						cluster_name       = "test-cluster"
						name               = "test-cluster.pool"
						replicas           = 2
						node_drain_grace_period ="5m"
						auto_repair        = true
						subnet_id          = "subnet-0abc123"
						instance_type      = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				// Update to 10m
				seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.pool", CombineHandlers(
					RespondWith(http.StatusOK, responseWith10m, jsonHeader),
				))
				seaplaneServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/api/v0/nodepools/test-cluster.pool"),
						verifyJQ(`.spec.nodePool.nodeDrainTimeout`, "10m"),
						RespondWith(http.StatusOK, responseWith10m, jsonHeader),
					),
					CombineHandlers(VerifyRequest(http.MethodDelete, "/api/v0/nodepools/test-cluster.pool"), RespondWith(http.StatusNoContent, "")),
				)

				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider           = rhcs.sp
						cluster_id         = "c-123"
						cluster_name       = "test-cluster"
						name               = "test-cluster.pool"
						replicas           = 2
						node_drain_grace_period ="10m"
						auto_repair        = true
						subnet_id          = "subnet-0abc123"
						instance_type      = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput = Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())

				runOutput = Terraform.Destroy()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("rejects invalid duration strings", func() {
				Terraform.Source(fmt.Sprintf(`
					provider "rhcs" {
						alias          = "sp"
						seaplane_url   = "%s"
						aws_account_id = "123"
						aws_region     = "us-east-1"
					}
					resource "rhcs_nodepool_seaplane" "test" {
						provider           = rhcs.sp
						cluster_id         = "c-123"
						cluster_name       = "test-cluster"
						name               = "test-cluster.pool"
						replicas           = 2
						node_drain_grace_period ="invalid"
						auto_repair        = true
						subnet_id          = "subnet-0abc123"
						instance_type      = "m5.xlarge"
					}
				`, seaplaneServer.URL()))

				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("must be a valid Go duration string")
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

	It("allows max_unavailable with autoscaling", func() {
		clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"nodePoolManagementARN": "arn:aws:iam::123456789012:role/roles-prefix-node-pool-management"}}}}}}`
		nodepoolResponse := `{"metadata": {"uid": "np-123", "name": "test-cluster.workers"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "autoScaling": {"min": 2, "max": 5}, "management": {"upgradeType": "Replace", "replace": {"strategy": "RollingUpdate", "rollingUpdate": {"maxUnavailable": 1}}}, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
		jsonHeader := http.Header{"Content-Type": []string{"application/json"}}

		seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.workers", CombineHandlers(
			RespondWith(http.StatusOK, nodepoolResponse, jsonHeader),
		))
		seaplaneServer.AppendHandlers(
			CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse, jsonHeader)),
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
				verifyJQ(`.spec.maxUnavailable`, "1"),
				verifyJQ(`.spec.nodePool.autoScaling.min`, 2),
				verifyJQ(`.spec.nodePool.autoScaling.max`, 5),
				RespondWith(http.StatusCreated, nodepoolResponse, jsonHeader),
			),
		)

		Terraform.Source(fmt.Sprintf(`
			provider "rhcs" {
				alias          = "sp"
				seaplane_url   = "%s"
				aws_account_id = "123"
				aws_region     = "us-east-1"
			}
			resource "rhcs_nodepool_seaplane" "test" {
				provider        = rhcs.sp
				cluster_id      = "c-123"
				cluster_name    = "test-cluster"
				name            = "test-cluster.workers"
				autoscaling     = {
					enabled      = true
					min_replicas = 2
					max_replicas = 5
				}
				max_unavailable = "1"
				auto_repair     = true
				subnet_id       = "subnet-0abc123"
				instance_type   = "m5.xlarge"
			}
		`, seaplaneServer.URL()))

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		planOutput := Terraform.Run("plan", "-detailed-exitcode")
		Expect(planOutput.ExitCode).To(BeZero(), "expected no-op plan after create with autoscaling + max_unavailable")

		runOutput = Terraform.Destroy()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("rejects max_unavailable exceeding autoscaling min_replicas", func() {
		Terraform.Source(fmt.Sprintf(`
			provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
			resource "rhcs_nodepool_seaplane" "test" {
				cluster_id      = "c-123"
				cluster_name    = "test-cluster"
				name            = "test-cluster.workers"
				autoscaling     = {
					enabled      = true
					min_replicas = 2
					max_replicas = 5
				}
				max_unavailable = "3"
				auto_repair     = true
				subnet_id       = "subnet-0abc123"
				instance_type   = "m5.xlarge"
			}
		`, seaplaneServer.URL()))

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("cannot exceed autoscaling min_replicas")
	})

	It("removes max_unavailable when unset", func() {
		clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"prefix": "roles-prefix"}}}}}}`
		createResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"maxUnavailable": "2", "autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`
		updateResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`

		seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.workers", CombineHandlers(
			RespondWith(http.StatusOK, createResponse, http.Header{"Content-Type": []string{"application/json"}}),
		))
		seaplaneServer.AppendHandlers(
			CombineHandlers(VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster"), RespondWith(http.StatusOK, clusterResponse)),
			CombineHandlers(VerifyRequest(http.MethodPost, "/api/v0/nodepools"), RespondWith(http.StatusCreated, createResponse)),
		)

		// Create with max_unavailable = "2"
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

		// Update to remove max_unavailable
		seaplaneServer.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.workers", CombineHandlers(
			RespondWith(http.StatusOK, updateResponse, http.Header{"Content-Type": []string{"application/json"}}),
		))
		seaplaneServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPut, "/api/v0/nodepools/test-cluster.workers"),
				VerifyJSONRepresenting(map[string]interface{}{
					"metadata": map[string]interface{}{"name": "test-cluster.workers"},
					"spec": map[string]interface{}{
						"autoRepair": true,
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
				RespondWith(http.StatusOK, updateResponse, http.Header{"Content-Type": []string{"application/json"}}),
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

	It("accepts max_unavailable = \"0\"", func() {
		clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"prefix": "roles-prefix"}}}}}}`
		createResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"maxUnavailable": "0", "autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`

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
						"maxUnavailable": "0",
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
				max_unavailable  = "0"
				auto_repair      = true
				subnet_id        = "subnet-0abc123"
				instance_type    = "m5.xlarge"
			}
		`, seaplaneServer.URL()))

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		planOutput := Terraform.Run("plan", "-detailed-exitcode")
		Expect(planOutput.ExitCode).To(BeZero(), "expected no-op plan after create with max_unavailable=0")

		runOutput = Terraform.Destroy()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	Context("Surge budget", func() {
		It("sets max_surge as an absolute count on create", func() {
			clusterResponse := `{"metadata": {"uid": "c-123", "name": "test-cluster"}, "spec": {"hostedCluster": {"platform": {"aws": {"rolesRef": {"prefix": "roles-prefix"}}}}}}`
			createResponse := `{"metadata": {"uid": "np-456", "name": "test-cluster.workers"}, "spec": {"maxSurge": "2", "autoRepair": true, "nodePool": {"clusterName": "test-cluster", "replicas": 5, "platform": {"type": "AWS", "aws": {"instanceType": "m5.xlarge", "instanceProfile": "roles-prefix-ROSA-Worker-Role", "subnet": {"id": "subnet-0abc123"}}}}},"status": {"phase": "Ready"}}`

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
							"maxSurge":   "2",
							"autoRepair": true,
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
					max_surge        = "2"
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

		It("rejects percentage max_surge with explanation", func() {
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
				resource "rhcs_nodepool_seaplane" "test" {
					cluster_id       = "c-123"
					cluster_name     = "test-cluster"
					name             = "test-cluster.workers"
					replicas         = 5
					max_surge        = "50%%"
					auto_repair      = true
					subnet_id        = "subnet-0abc123"
					instance_type    = "m5.xlarge"
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("percentages are not supported yet")
		})

		It("rejects non-integer max_surge", func() {
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" { seaplane_url = "%s"; aws_account_id = "123"; aws_region = "us-east-1" }
				resource "rhcs_nodepool_seaplane" "test" {
					cluster_id       = "c-123"
					cluster_name     = "test-cluster"
					name             = "test-cluster.workers"
					replicas         = 5
					max_surge        = "abc"
					auto_repair      = true
					subnet_id        = "subnet-0abc123"
					instance_type    = "m5.xlarge"
				}
			`, seaplaneServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("must be a non-negative whole number")
		})
	})
})
})

var _ = Describe("Seaplane nodepool root volume", func() {
	DescribeTable("plans the root volume lifecycle", func(nextSize string) {
		server := NewServer()
		DeferCleanup(server.Close)
		header := http.Header{"Content-Type": []string{"application/json"}}
		cluster := `{"metadata":{"name":"test-cluster","uid":"cluster-uid"},"spec":{"hostedCluster":{"platform":{"aws":{"rolesRef":{"nodePoolManagementARN":"arn:aws:iam::123456789012:role/test-NodePool"}}}}}}`
		// Include service-managed fields to exercise refresh without a spurious diff.
		pool := `{"metadata":{"name":"test-cluster.worker","uid":"pool-uid"},"spec":{"nodePool":{"clusterName":"test-cluster","replicas":1,"platform":{"type":"AWS","aws":{"instanceType":"m5.xlarge","rootVolume":{"size":500,"type":"gp3","encrypted":true}}}}},"status":{"phase":"Ready"}}`
		server.RouteToHandler(http.MethodGet, "/api/v0/clusters/test-cluster", RespondWith(http.StatusOK, cluster, header))
		server.RouteToHandler(http.MethodGet, "/api/v0/nodepools/test-cluster.worker", RespondWith(http.StatusOK, pool, header))
		server.AppendHandlers(CombineHandlers(
			VerifyRequest(http.MethodPost, "/api/v0/nodepools"),
			RespondWith(http.StatusCreated, pool, header),
		))
		source := func(size string) string {
			return fmt.Sprintf(`
                provider "rhcs" {
                    alias = "sp"
                    seaplane_url = %q
                    aws_account_id = "123456789012"
                    aws_region = "us-east-1"
                }
                resource "rhcs_nodepool_seaplane" "test" {
                    provider = rhcs.sp
                    name = "test-cluster.worker"
                    cluster_name = "test-cluster"
                    cluster_id = "cluster-uid"
                    replicas = 1
                    instance_type = "m5.xlarge"
                    %s
                }
            `, server.URL(), size)
		}
		Terraform.Source(source("disk_size = 500"))
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		Expect(Terraform.Run("plan", "-detailed-exitcode", "-no-color").ExitCode).To(BeZero())
		Terraform.Source(source(nextSize))
		plan := Terraform.Run("plan", "-detailed-exitcode", "-no-color")
		Expect(plan.ExitCode).To(Equal(2))
		plan.VerifyOutputContainsSubstring("must be replaced")
		plan.VerifyOutputContainsSubstring("disk_size")
		plan.VerifyOutputContainsSubstring("forces replacement")
	},
		Entry("replaces the pool when changing size", "disk_size = 600"),
		Entry("replaces the pool when removing explicit size", ""),
	)
})
