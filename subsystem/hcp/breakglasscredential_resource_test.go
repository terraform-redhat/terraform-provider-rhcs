/*
Copyright (c) 2025 Red Hat, Inc.

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
	"time"

	. "github.com/onsi/ginkgo/v2"    // nolint
	. "github.com/onsi/gomega"       // nolint
	. "github.com/onsi/gomega/ghttp" // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("Break Glass Credential", func() {
	Context("creation", func() {
		It("fails if cluster is not HCP", func() {
			cluster, err := cmv1.NewCluster().
				ID("123").
				Name("cluster").
				Hypershift(cmv1.NewHypershift().Enabled(false)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithOcmObjectMarshal(http.StatusOK, cluster, cmv1.MarshalCluster),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Break glass credentials are only supported on Hosted Control Plane clusters")
		})

		It("fails if external authentication is not enabled", func() {
			cluster, err := cmv1.NewCluster().
				ID("123").
				Name("cluster").
				Hypershift(cmv1.NewHypershift().Enabled(true)).
				ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(false)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithOcmObjectMarshal(http.StatusOK, cluster, cmv1.MarshalCluster),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("External Authentication Configuration is not enabled")
		})

		It("fails if username does not match regex pattern", func() {
			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					username = "invalid@user!"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("must respect the regexp")
		})

		It("fails if username exceeds maximum length", func() {
			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					username = "this-is-a-very-long-username-that-exceeds-the-maximum-allowed-length"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("string length must be at most 35")
		})

		It("fails if expiration duration is less than 10 minutes", func() {
			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					expiration_duration = "5m"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("The expiration duration needs to be at least 10 minutes")
		})

		It("fails if expiration duration exceeds 24 hours", func() {
			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					expiration_duration = "25h"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("The expiration duration needs to be at maximum 24 hours")
		})

		It("fails if expiration duration has invalid format", func() {
			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					expiration_duration = "invalid"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Invalid configuration")
		})

		It("successfully creates a break glass credential with default settings", func() {
			expirationTime := time.Now().Add(1 * time.Hour)

			cluster, err := cmv1.NewCluster().
				ID("123").
				Name("cluster").
				Hypershift(cmv1.NewHypershift().Enabled(true)).
				ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(true)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			breakGlassCreate, err := cmv1.NewBreakGlassCredential().
				ID("bgc-123").
				Username("break-glass-user").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Build()
			Expect(err).ToNot(HaveOccurred())

			breakGlassWithKubeconfig, err := cmv1.NewBreakGlassCredential().
				ID("bgc-123").
				Username("break-glass-user").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Kubeconfig("apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: https://api.example.com\n  name: break-glass").
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithOcmObjectMarshal(http.StatusOK, cluster, cmv1.MarshalCluster),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials"),
					RespondWithOcmObjectMarshal(http.StatusCreated, breakGlassCreate, cmv1.MarshalBreakGlassCredential),
				),
				// Polling for kubeconfig
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-123"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlassWithKubeconfig, cmv1.MarshalBreakGlassCredential),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_break_glass_credential", "break_glass").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			attributes := actualResource["attributes"].(map[string]interface{})
			Expect(attributes["cluster"]).To(Equal("123"))
			Expect(attributes["id"]).To(Equal("bgc-123"))
			Expect(attributes["username"]).To(Equal("break-glass-user"))
			Expect(attributes["status"]).To(Equal("issued"))
			Expect(attributes["kubeconfig"]).ToNot(BeEmpty())
		})

		It("successfully creates a break glass credential with custom username", func() {
			expirationTime := time.Now().Add(1 * time.Hour)

			cluster, err := cmv1.NewCluster().
				ID("123").
				Name("cluster").
				Hypershift(cmv1.NewHypershift().Enabled(true)).
				ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(true)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			breakGlassCreate, err := cmv1.NewBreakGlassCredential().
				ID("bgc-456").
				Username("custom-admin").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Build()
			Expect(err).ToNot(HaveOccurred())

			breakGlassWithKubeconfig, err := cmv1.NewBreakGlassCredential().
				ID("bgc-456").
				Username("custom-admin").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Kubeconfig("apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: https://api.example.com\n  name: break-glass").
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithOcmObjectMarshal(http.StatusOK, cluster, cmv1.MarshalCluster),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials"),
					VerifyJQ(".username", "custom-admin"),
					RespondWithOcmObjectMarshal(http.StatusCreated, breakGlassCreate, cmv1.MarshalBreakGlassCredential),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-456"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlassWithKubeconfig, cmv1.MarshalBreakGlassCredential),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					username = "custom-admin"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_break_glass_credential", "break_glass").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			attributes := actualResource["attributes"].(map[string]interface{})
			Expect(attributes["username"]).To(Equal("custom-admin"))
		})

		It("successfully creates a break glass credential with custom expiration duration", func() {
			expirationTime := time.Now().Add(2 * time.Hour)

			cluster, err := cmv1.NewCluster().
				ID("123").
				Name("cluster").
				Hypershift(cmv1.NewHypershift().Enabled(true)).
				ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(true)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			breakGlassCreate, err := cmv1.NewBreakGlassCredential().
				ID("bgc-789").
				Username("break-glass-user").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Build()
			Expect(err).ToNot(HaveOccurred())

			breakGlassWithKubeconfig, err := cmv1.NewBreakGlassCredential().
				ID("bgc-789").
				Username("break-glass-user").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Kubeconfig("apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: https://api.example.com\n  name: break-glass").
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithOcmObjectMarshal(http.StatusOK, cluster, cmv1.MarshalCluster),
				),
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials"),
					RespondWithOcmObjectMarshal(http.StatusCreated, breakGlassCreate, cmv1.MarshalBreakGlassCredential),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-789"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlassWithKubeconfig, cmv1.MarshalBreakGlassCredential),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
					expiration_duration = "2h"
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_break_glass_credential", "break_glass").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			attributes := actualResource["attributes"].(map[string]interface{})
			Expect(attributes["expiration_duration"]).To(Equal("2h"))
			Expect(attributes["expiration_timestamp"]).ToNot(BeEmpty())
		})
	})

	Context("reading", func() {
		It("successfully reads break glass credential state", func() {
			expirationTime := time.Now().Add(1 * time.Hour)

			breakGlass, err := cmv1.NewBreakGlassCredential().
				ID("bgc-123").
				Username("break-glass-user").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Kubeconfig("apiVersion: v1\nkind: Config").
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-123"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlass, cmv1.MarshalBreakGlassCredential),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-123"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlass, cmv1.MarshalBreakGlassCredential),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Import("rhcs_break_glass_credential.break_glass", "123,bgc-123")
			Expect(runOutput.ExitCode).To(BeZero())
		})
	})

	Context("importing", func() {
		It("fails if import identifier is invalid", func() {
			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Import("rhcs_break_glass_credential.break_glass", "invalid-id")
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Invalid import identifier")
		})

		It("successfully imports break glass credential with correct format", func() {
			expirationTime := time.Now().Add(1 * time.Hour)

			breakGlass, err := cmv1.NewBreakGlassCredential().
				ID("bgc-123").
				Username("imported-user").
				Status(cmv1.BreakGlassCredentialStatusIssued).
				ExpirationTimestamp(expirationTime).
				Kubeconfig("apiVersion: v1\nkind: Config").
				Build()
			Expect(err).ToNot(HaveOccurred())

			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-123"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlass, cmv1.MarshalBreakGlassCredential),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/break_glass_credentials/bgc-123"),
					RespondWithOcmObjectMarshal(http.StatusOK, breakGlass, cmv1.MarshalBreakGlassCredential),
				),
			)

			Terraform.Source(`
				resource "rhcs_break_glass_credential" "break_glass" {
					cluster = "123"
				}
			`)
			runOutput := Terraform.Import("rhcs_break_glass_credential.break_glass", "123,bgc-123")
			Expect(runOutput.ExitCode).To(BeZero())

			actualResource, ok := Terraform.Resource("rhcs_break_glass_credential", "break_glass").(map[string]interface{})
			Expect(ok).To(BeTrue(), "Type conversion failed for the received resource state")

			attributes := actualResource["attributes"].(map[string]interface{})
			Expect(attributes["cluster"]).To(Equal("123"))
			Expect(attributes["id"]).To(Equal("bgc-123"))
			Expect(attributes["username"]).To(Equal("imported-user"))
		})
	})
})
