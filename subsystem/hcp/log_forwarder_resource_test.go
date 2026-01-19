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
	"net/http"

	. "github.com/onsi/ginkgo/v2"                      // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("Log Forwarder Resource", func() {
	clusterReady := `{
		"kind": "Cluster",
		"id": "123",
		"href": "/api/clusters_mgmt/v1/clusters/123",
		"name": "test-cluster",
		"state": "ready"
	}`

	logForwarderS3Response := `{
		"kind": "LogForwarder",
		"id": "log-fwd-1",
		"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
		"s3": {
			"bucket_name": "my-logs",
			"bucket_prefix": "rosa/"
		},
		"applications": ["audit", "infrastructure"]
	}`

	logForwarderCloudWatchResponse := `{
		"kind": "LogForwarder",
		"id": "log-fwd-2",
		"href": "/api/clusters_mgmt/v1/clusters/123/log_forwarders/log-fwd-2",
		"cloudwatch": {
			"log_group_name": "/aws/rosa/logs",
			"log_distribution_role_arn": "arn:aws:iam::123456789012:role/log-forwarder"
		},
		"applications": ["audit"]
	}`

	logForwarderSingleAppResponse := `{
		"kind": "LogForwarder",
		"id": "log-fwd-1",
		"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
		"s3": {
			"bucket_name": "my-logs",
			"bucket_prefix": ""
		},
		"applications": ["audit"]
	}`

	Context("validation", func() {
		It("fails if cluster ID is empty", func() {
			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = ""
					s3 = {
						bucket_name = "my-logs"
					}
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("cluster ID may not be empty/blank string")
		})

		It("fails if neither S3 nor CloudWatch is configured", func() {
			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
		})

		It("fails if both S3 and CloudWatch are configured", func() {
			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
					}
					cloudwatch = {
						log_group_name = "/aws/logs"
						log_distribution_role_arn = "arn:aws:iam::123:role/log"
					}
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("one (and only one)")
		})

		It("fails if S3 bucket_name is empty", func() {
			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = ""
					}
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
		})

		It("fails if neither applications nor groups are specified", func() {
			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
					}
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("At least one of 'applications' or 'groups'")
		})

		It("fails when attempting to change immutable cluster field", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReady),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					RespondWithJSON(http.StatusCreated, logForwarderSingleAppResponse),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, logForwarderSingleAppResponse),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
					}
					applications = ["audit"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "456"
					s3 = {
						bucket_name = "my-logs"
					}
					applications = ["audit"]
				}
			`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute value cannot be changed")
			runOutput.VerifyErrorContainsSubstring("Attribute cluster, cannot be changed from")
		})
	})

	Context("create", func() {
		BeforeEach(func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReady),
				),
			)
		})

		It("creates log forwarder with S3 configuration", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					VerifyJQ(".s3.bucket_name", "my-logs"),
					VerifyJQ(".s3.bucket_prefix", "rosa/"),
					VerifyJQ(".applications[0]", "audit"),
					VerifyJQ(".applications[1]", "infrastructure"),
					RespondWithJSON(http.StatusCreated, logForwarderS3Response),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
						bucket_prefix = "rosa/"
					}
					applications = ["audit", "infrastructure"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("creates log forwarder with CloudWatch configuration", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					VerifyJQ(".cloudwatch.log_group_name", "/aws/rosa/logs"),
					VerifyJQ(".cloudwatch.log_distribution_role_arn", "arn:aws:iam::123456789012:role/log-forwarder"),
					VerifyJQ(".applications[0]", "audit"),
					RespondWithJSON(http.StatusCreated, logForwarderCloudWatchResponse),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					cloudwatch = {
						log_group_name = "/aws/rosa/logs"
						log_distribution_role_arn = "arn:aws:iam::123456789012:role/log-forwarder"
					}
					applications = ["audit"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("creates log forwarder with both applications and groups", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					VerifyJQ(".s3.bucket_name", "my-logs"),
					VerifyJQ(".applications[0]", "audit"),
					VerifyJQ(".groups[0].id", "group-1"),
					VerifyJQ(".groups[0].version", "1.0.0"),
					RespondWithJSON(http.StatusCreated, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs",
							"bucket_prefix": ""
						},
						"applications": ["audit"],
						"groups": [
							{
								"id": "group-1",
								"version": "1.0.0"
							}
						]
					}`),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
					}
					applications = ["audit"]
					groups = [
						{
							id = "group-1"
							version = "1.0.0"
						}
					]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("creates log forwarder with empty bucket_prefix", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					VerifyJQ(".s3.bucket_name", "my-logs"),
					VerifyJQ(".s3.bucket_prefix", ""),
					VerifyJQ(".applications[0]", "audit"),
					RespondWithJSON(http.StatusCreated, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs",
							"bucket_prefix": ""
						},
						"applications": ["audit"]
					}`),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
						bucket_prefix = ""
					}
					applications = ["audit"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("update", func() {
		BeforeEach(func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReady),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					RespondWithJSON(http.StatusCreated, logForwarderS3Response),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, logForwarderS3Response),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
						bucket_prefix = "rosa/"
					}
					applications = ["audit", "infrastructure"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates S3 bucket configuration", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					VerifyJQ(".s3.bucket_name", "my-logs-updated"),
					VerifyJQ(".s3.bucket_prefix", "logs/"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs-updated",
							"bucket_prefix": "logs/"
						},
						"applications": ["audit", "infrastructure"]
					}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs-updated",
							"bucket_prefix": "logs/"
						},
						"applications": ["audit", "infrastructure"]
					}`),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs-updated"
						bucket_prefix = "logs/"
					}
					applications = ["audit", "infrastructure"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("updates applications list", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs",
							"bucket_prefix": "rosa/"
						},
						"applications": ["audit", "infrastructure", "application"]
					}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs",
							"bucket_prefix": "rosa/"
						},
						"applications": ["audit", "infrastructure", "application"]
					}`),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
						bucket_prefix = "rosa/"
					}
					applications = ["audit", "infrastructure", "application"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("delete", func() {
		BeforeEach(func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterReady),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders"),
					RespondWithJSON(http.StatusCreated, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs",
							"bucket_prefix": "rosa/"
						},
						"applications": ["audit"]
					}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "LogForwarder",
						"id": "log-fwd-1",
						"href": "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1",
						"s3": {
							"bucket_name": "my-logs",
							"bucket_prefix": "rosa/"
						},
						"applications": ["audit"]
					}`),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
						bucket_prefix = "rosa/"
					}
					applications = ["audit"]
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("deletes log forwarder successfully", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusNoContent, "{}"),
				),
			)

			// Remove resource from configuration
			Terraform.Source("")
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("import", func() {
		It("imports existing log forwarder", func() {
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, logForwarderS3Response),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/control_plane/log_forwarders/log-fwd-1"),
					RespondWithJSON(http.StatusOK, logForwarderS3Response),
				),
			)

			Terraform.Source(`
				resource "rhcs_log_forwarder" "log_forwarder" {
					cluster = "123"
					s3 = {
						bucket_name = "my-logs"
						bucket_prefix = "rosa/"
					}
					applications = ["audit", "infrastructure"]
				}
			`)
			runOutput := Terraform.Import("rhcs_log_forwarder.log_forwarder", "123,log-fwd-1")
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})
})
