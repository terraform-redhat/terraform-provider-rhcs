// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("rhcs_dns_domain_seaplane", func() {
	const (
		reservationName   = "reservation-a"
		reservationUID    = "c8e1b8db-87f0-4a83-922d-ff5ab5d81f9f"
		reservationDomain = "f7a3.0.openshiftapps.com"
	)

	reservationResponse := func(phase, baseDomain string) string {
		status := `{"phase":"` + phase + `"}`
		if baseDomain != "" {
			status = `{"phase":"` + phase + `","baseDomain":"` + baseDomain + `"}`
		}
		return `{
			"kind": "DNSReservation",
			"apiVersion": "seaplane.io/v1alpha1",
			"metadata": {"name": "` + reservationName + `", "uid": "` + reservationUID + `"},
			"status": ` + status + `
		}`
	}
	readyResponse := reservationResponse("Ready", reservationDomain)
	header := http.Header{"Content-Type": []string{"application/json"}}
	verifyCreate := func(_ http.ResponseWriter, req *http.Request) {
		var body struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}
		Expect(json.NewDecoder(req.Body).Decode(&body)).To(Succeed())
		Expect(body.Metadata.Name).To(Equal(reservationName))
	}

	It("creates, refreshes, imports, and deletes a DNS reservation", func() {
		server := NewServer()
		DeferCleanup(server.Close)

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/v0/dns_reservations"),
				verifyCreate,
				RespondWith(http.StatusCreated, reservationResponse("Pending", ""), header),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusOK, readyResponse, header),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusOK, readyResponse, header),
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
				name     = "` + reservationName + `"
			}
		`)

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())
		stateResource := Terraform.Resource("rhcs_dns_domain_seaplane", "test")
		Expect(stateResource).To(matchJQ(".attributes.name", reservationName))
		Expect(stateResource).To(matchJQ(".attributes.id", reservationUID))
		Expect(stateResource).To(matchJQ(".attributes.base_domain", reservationDomain))
		Expect(stateResource).To(matchJQ(".attributes.phase", "Ready"))

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusOK, readyResponse, header),
			),
		)

		runOutput = Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusOK, readyResponse, header),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusAccepted, nil, header),
			),
		)

		runOutput = Terraform.Destroy()
		Expect(runOutput.ExitCode).To(BeZero())
	})

	It("imports by reservation name and stores the UID as id", func() {
		server := NewServer()
		DeferCleanup(server.Close)

		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusOK, readyResponse, header),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/v0/dns_reservations/"+reservationName),
				RespondWith(http.StatusOK, readyResponse, header),
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
				name     = "` + reservationName + `"
			}
		`)

		runOutput := Terraform.Import("rhcs_dns_domain_seaplane.test", reservationName)
		Expect(runOutput.ExitCode).To(BeZero())
		stateResource := Terraform.Resource("rhcs_dns_domain_seaplane", "test")
		Expect(stateResource).To(matchJQ(".attributes.name", reservationName))
		Expect(stateResource).To(matchJQ(".attributes.id", reservationUID))
		Expect(stateResource).To(matchJQ(".attributes.base_domain", reservationDomain))
	})

	It("rejects an empty reservation name before calling the API", func() {
		Terraform.Source(`
			provider "rhcs" {
				alias          = "sp"
				seaplane_url   = "https://example.com"
				aws_account_id = "123456789012"
				aws_region     = "us-east-1"
			}

			resource "rhcs_dns_domain_seaplane" "test" {
				provider = rhcs.sp
				name     = ""
			}
		`)

		runOutput := Terraform.Run("plan", "-no-color")
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("string length must be at least 1")
	})

	It("returns an actionable error without Seaplane configuration", func() {
		Terraform.Source(`
			resource "rhcs_dns_domain_seaplane" "test" {
				name = "reservation-a"
			}
		`)

		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("seaplane_url")
	})
})
