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

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_cluster_hyperfleet", func() {
	Context("Provider configuration validation", func() {
		It("Reports a descriptive error when hyperfleet_url is absent from the provider", func() {
			Terraform.Source(`
				resource "rhcs_cluster_hyperfleet" "test" {
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("hyperfleet_url")
		})

		It("Reports a descriptive error when aws_account_id is absent", func() {
			Terraform.Source(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("aws_account_id")
		})

		It("Reports a descriptive error when aws_region contradicts the region in hyperfleet_url", func() {
			Terraform.Source(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "https://abc123.execute-api.us-east-1.amazonaws.com/prod"
					aws_account_id = "123456789012"
					aws_region     = "us-west-2"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
				}
			`)
			runOutput := Terraform.Apply()
			runOutput.VerifyErrorContainsSubstring("AWS region mismatch")
		})
	})

	Context("Create operation", func() {
		var hyperfleetServer *Server

		BeforeEach(func() {
			hyperfleetServer = NewServer()
		})

		AfterEach(func() {
			hyperfleetServer.Close()
		})

		It("creates a cluster successfully", func() {
			clusterResponse := `{
				"id": "test-cluster-id",
				"name": "my-cluster",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1"
							}
						}
					}
				},
				"status": {
					"phase": "Provisioning",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, clusterResponse, header),
				),
			)

			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
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

		It("updates cluster expiration_timestamp", func() {
			clusterWithoutExpiration := `{
				"id": "test-cluster-id",
				"name": "my-cluster",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1"
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			clusterWithExpiration := `{
				"id": "test-cluster-id",
				"name": "my-cluster",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
				"spec": {
					"expirationTimestamp": "2025-01-01T00:00:00Z",
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1"
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				// First apply: create without expiration
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, clusterWithoutExpiration, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster-id"),
					RespondWith(http.StatusOK, clusterWithoutExpiration, header),
				),
				// Second apply: update to add expiration
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster-id"),
					RespondWith(http.StatusOK, clusterWithoutExpiration, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPut, "/api/v0/clusters/test-cluster-id"),
					RespondWith(http.StatusOK, clusterWithExpiration, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster-id"),
					RespondWith(http.StatusOK, clusterWithExpiration, header),
				),
			)

			// First apply: create without expiration
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Second apply: update to add expiration
			Terraform.Source(fmt.Sprintf(`
				provider "rhcs" {
					alias          = "hf"
					hyperfleet_url = "%s"
					aws_account_id = "123456789012"
					aws_region     = "us-east-1"
				}

				resource "rhcs_cluster_hyperfleet" "test" {
					provider                = rhcs.hf
					name                    = "my-cluster"
					operator_roles_prefix   = "my-cluster"
					aws_subnet_ids         = ["subnet-0abc123"]
					vpc_id                 = "vpc-0def456"
					availability_zones     = ["us-east-1a"]
					expiration_timestamp   = "2025-01-01T00:00:00Z"
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

		It("deletes a cluster successfully", func() {
			createResponse := `{
				"id": "test-cluster-id",
				"name": "my-cluster",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
				"spec": {
					"hostedCluster": {
						"platform": {
							"type": "AWS",
							"aws": {
								"region": "us-east-1"
							}
						}
					}
				},
				"status": {
					"phase": "Ready",
					"controlPlaneEndpoint": {
						"host": "api.my-cluster.example.com",
						"port": 6443
					}
				}
			}`

			header := http.Header{"Content-Type": []string{"application/json"}}

			hyperfleetServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/v0/clusters"),
					RespondWith(http.StatusCreated, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/v0/clusters/test-cluster-id"),
					RespondWith(http.StatusOK, createResponse, header),
				),
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/v0/clusters/test-cluster-id"),
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

				resource "rhcs_cluster_hyperfleet" "test" {
					provider              = rhcs.hf
					name                  = "my-cluster"
					operator_roles_prefix = "my-cluster"
					aws_subnet_ids       = ["subnet-0abc123"]
					vpc_id               = "vpc-0def456"
					availability_zones   = ["us-east-1a"]
				}
			`, hyperfleetServer.URL()))

			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			runOutput = Terraform.Destroy()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})
})
