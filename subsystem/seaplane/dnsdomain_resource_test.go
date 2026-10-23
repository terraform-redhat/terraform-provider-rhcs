// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_dns_domain_seaplane", func() {
	const domainID = "abcd.0.example.com"

	domainResponse := func(id string) string {
		return `{
			"kind": "DNSDomain",
			"id": "` + id + `",
			"cluster_arch": "hcp",
			"user_defined": true,
			"reserved_at_timestamp": "2026-10-08T12:00:00Z"
		}`
	}
	listResponse := func(id string) string {
		return `{
			"kind": "DNSDomainList",
			"page": 1,
			"size": 1,
			"total": 1,
			"items": [` + domainResponse(id) + `]
		}`
	}

	It("creates, refreshes, and deletes an HCP DNS domain", func() {
		server := NewServer()
		DeferCleanup(server.Close)

		header := http.Header{"Content-Type": []string{"application/json"}}
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/v0/dns_domains"),
				VerifyJSON(`{"cluster_arch":"hcp"}`),
				RespondWith(http.StatusCreated, domainResponse(domainID), header),
			),
		)

		Terraform.Source(`
			provider "rhcs" {
				alias          = "sp"
				seaplane_url   = "` + server.URL() + `"
				aws_account_id = "123456789012"
				aws_region     = "us-east-1"
			}

			resource "rhcs_dns_domain_seaplane" "test" {
				provider     = rhcs.sp
				cluster_arch = "hcp"
			}
		`)

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
		stateResource := Terraform.Resource("rhcs_dns_domain_seaplane", "test")
		Expect(stateResource).To(matchJQ(".attributes.id", domainID))
		Expect(stateResource).To(matchJQ(".attributes.cluster_arch", "hcp"))

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_domains"),
				RespondWith(http.StatusOK, listResponse(domainID), header),
			),
		)

		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_domains"),
				RespondWith(http.StatusOK, listResponse(domainID), header),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/v0/dns_domains/"+domainID),
				RespondWith(http.StatusNoContent, nil, header),
			),
		)

		runOutput = Terraform.Destroy()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("imports an existing DNS domain", func() {
		server := NewServer()
		DeferCleanup(server.Close)

		header := http.Header{"Content-Type": []string{"application/json"}}
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_domains"),
				RespondWith(http.StatusOK, listResponse(domainID), header),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_domains"),
				RespondWith(http.StatusOK, listResponse(domainID), header),
			),
		)

		Terraform.Source(`
			provider "rhcs" {
				alias          = "sp"
				seaplane_url   = "` + server.URL() + `"
				aws_account_id = "123456789012"
				aws_region     = "us-east-1"
			}

			resource "rhcs_dns_domain_seaplane" "test" {
				provider = rhcs.sp
			}
		`)

		runOutput := Terraform.Import("rhcs_dns_domain_seaplane.test", domainID)
		Expect(runOutput.ExitCode).To(BeZero())
		stateResource := Terraform.Resource("rhcs_dns_domain_seaplane", "test")
		Expect(stateResource).To(matchJQ(".attributes.id", domainID))
		Expect(stateResource).To(matchJQ(".attributes.cluster_arch", "hcp"))
	})

	It("rejects a non-HCP architecture before calling the API", func() {
		Terraform.Source(`
			provider "rhcs" {
				alias          = "sp"
				seaplane_url   = "https://example.com"
				aws_account_id = "123456789012"
				aws_region     = "us-east-1"
			}

			resource "rhcs_dns_domain_seaplane" "test" {
				provider     = rhcs.sp
				cluster_arch = "classic"
			}
		`)

		runOutput := Terraform.Run("plan", "-no-color")
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("value must be one of")
	})

	It("returns an actionable error without Seaplane configuration", func() {
		Terraform.Source(`
			resource "rhcs_dns_domain_seaplane" "test" {}
		`)

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("seaplane_url")
	})
})
