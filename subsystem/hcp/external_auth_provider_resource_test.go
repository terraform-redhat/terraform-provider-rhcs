// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"encoding/json"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"    // nolint
	. "github.com/onsi/gomega"       // nolint
	. "github.com/onsi/gomega/ghttp" // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

const externalAuthPath = "/api/clusters_mgmt/v1/clusters/123/external_auth_config/external_auths"

const externalAuthConfig = `
resource "rhcs_external_auth_provider" "example" {
  cluster                      = "123"
  name                         = "example"
  issuer_url                   = "https://issuer.example.com"
  issuer_audiences             = ["console", "other"]
  claim_validation_rule        = ["aud:urn:example:scope"]
  console_client_id            = "console"
  console_client_secret        = "fixture-placeholder"
  extra_scopes                 = ["profile"]
}
`

const externalAuthConfigUpdated = `
resource "rhcs_external_auth_provider" "example" {
  cluster                      = "123"
  name                         = "example"
  issuer_url                   = "https://issuer-new.example.com"
  issuer_audiences             = ["console", "other"]
  claim_validation_rule        = []
  console_client_id            = "console"
  console_client_secret        = "replacement-placeholder"
  extra_scopes                 = []
}
`

func externalAuthCluster(state cmv1.ClusterState, hcpEnabled, externalAuthEnabled bool) string {
	cluster, err := cmv1.NewCluster().ID("123").State(state).
		Hypershift(cmv1.NewHypershift().Enabled(hcpEnabled)).
		ExternalAuthConfig(cmv1.NewExternalAuthConfig().Enabled(externalAuthEnabled)).Build()
	Expect(err).ToNot(HaveOccurred())
	var body strings.Builder
	Expect(cmv1.MarshalCluster(cluster, &body)).To(Succeed())
	return body.String()
}

var _ = Describe("External Auth Provider Resource", func() {
	It("creates, refreshes, updates with minimal PATCH, and deletes", func() {
		clusterJSON := externalAuthCluster(cmv1.ClusterStateReady, true, true)
		currentJSON := `{
          "id":"example", "issuer":{"url":"https://issuer.example.com","audiences":["other","console"]},
          "claim":{"mappings":{"groups":{"claim":"groups","prefix":"team-"},
                   "username":{"claim":"email","prefix":"user-","prefix_policy":"Prefix"}},
                   "validation_rules":[{"claim":"aud","required_value":"urn:example:scope"}]},
          "clients":[{"id":"console","component":{"name":"console","namespace":"openshift-console"},
                      "secret":"********","extra_scopes":["profile"],"type":"confidential"}]
        }`
		var postBody, patchBody map[string]any
		var handlerErr error
		deleted := false
		TestServer.RouteToHandler(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, handlerErr = w.Write([]byte(clusterJSON))
		})
		TestServer.RouteToHandler(http.MethodGet, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				_, handlerErr = w.Write([]byte(`{"message":"not found"}`))
				return
			}
			_, handlerErr = w.Write([]byte(currentJSON))
		})
		TestServer.RouteToHandler(http.MethodPost, externalAuthPath, func(w http.ResponseWriter, r *http.Request) {
			handlerErr = json.NewDecoder(r.Body).Decode(&postBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(currentJSON))
		})
		TestServer.RouteToHandler(http.MethodPatch, externalAuthPath+"/example", func(w http.ResponseWriter, r *http.Request) {
			handlerErr = json.NewDecoder(r.Body).Decode(&patchBody)
			currentJSON = `{
              "id":"example", "issuer":{"url":"https://issuer-new.example.com","audiences":["other","console"]},
              "claim":{"mappings":{"groups":{"claim":"groups","prefix":"team-"},
                       "username":{"claim":"email","prefix":"user-","prefix_policy":"Prefix"}},
                       "validation_rules":[]},
              "clients":[{"id":"console","component":{"name":"console","namespace":"openshift-console"},
                          "extra_scopes":[],"type":"confidential"}]
            }`
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(currentJSON))
		})
		TestServer.RouteToHandler(http.MethodDelete, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		})

		Terraform.Source(externalAuthConfig)
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		Expect(handlerErr).ToNot(HaveOccurred())
		Expect(postBody).ToNot(BeNil())
		Expect(postBody).To(MatchJQ(`.issuer.url`, "https://issuer.example.com"))
		Expect(postBody).To(MatchJQ(`.clients[0].component.namespace`, "openshift-console"))
		Expect(postBody).To(MatchJQ(`.clients[0].secret`, "fixture-placeholder"))
		Expect(postBody).To(MatchJQ(`.claim.validation_rules[0].claim`, "aud"))
		Expect(postBody).To(MatchJQ(`.claim.validation_rules[0].required_value`, "urn:example:scope"))
		Expect(postBody["clients"].([]any)[0].(map[string]any)).ToNot(HaveKey("type"))
		Expect(Terraform.Resource("rhcs_external_auth_provider", "example")).To(MatchJQ(`.attributes.id`, "example"))
		Expect(Terraform.Run("refresh").ExitCode).To(BeZero())
		Expect(Terraform.Run("plan", "-detailed-exitcode").ExitCode).To(BeZero())
		Expect(Terraform.Resource("rhcs_external_auth_provider", "example")).To(
			MatchJQ(`.attributes.claim_validation_rule[0]`, "aud:urn:example:scope"),
		)

		Terraform.Source(externalAuthConfigUpdated)
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		Expect(handlerErr).ToNot(HaveOccurred())
		Expect(patchBody).ToNot(BeNil())
		Expect(patchBody).ToNot(HaveKey("id"))
		Expect(patchBody).To(MatchJQ(`.issuer.url`, "https://issuer-new.example.com"))
		Expect(patchBody).To(MatchJQ(`.clients[0].secret`, "replacement-placeholder"))
		Expect(patchBody).To(MatchJQ(`.clients[0].extra_scopes`, []any{}))
		Expect(patchBody).To(MatchJQ(`.claim.validation_rules`, []any{}))
		Expect(patchBody).To(MatchJQ(`.claim.mappings.groups.prefix`, "team-"))
		Expect(patchBody).To(MatchJQ(`.claim.mappings.username.prefix`, "user-"))
		Expect(patchBody).To(MatchJQ(`.claim.mappings.username.prefix_policy`, "Prefix"))
		Expect(Terraform.Run("plan", "-detailed-exitcode").ExitCode).To(BeZero())

		Terraform.Source("")
		Expect(Terraform.Destroy().ExitCode).To(BeZero())
		Expect(deleted).To(BeTrue())
	})

	It("retries pending deletion on PATCH but not other bad requests", func() {
		clusterJSON := externalAuthCluster(cmv1.ClusterStateReady, true, true)
		currentJSON := `{
          "id":"example", "issuer":{"url":"https://issuer.example.com","audiences":["other","console"]},
          "claim":{"mappings":{"groups":{"claim":"groups"},"username":{"claim":"email"}},
                   "validation_rules":[{"claim":"aud","required_value":"urn:example:scope"}]},
          "clients":[{"id":"console","component":{"name":"console","namespace":"openshift-console"},
                      "secret":"********","extra_scopes":["profile"],"type":"confidential"}]
        }`
		var patchBodies []map[string]any
		var handlerErr error
		TestServer.RouteToHandler(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, handlerErr = w.Write([]byte(clusterJSON))
		})
		TestServer.RouteToHandler(http.MethodPost, externalAuthPath, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, handlerErr = w.Write([]byte(currentJSON))
		})
		TestServer.RouteToHandler(http.MethodGet, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, handlerErr = w.Write([]byte(currentJSON))
		})
		TestServer.RouteToHandler(http.MethodPatch, externalAuthPath+"/example", func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			handlerErr = json.NewDecoder(r.Body).Decode(&body)
			patchBodies = append(patchBodies, body)
			w.Header().Set("Content-Type", "application/json")
			switch len(patchBodies) {
			case 1:
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"kind":"Error","id":"400",` +
					`"reason":"Cluster '123' has an external authentication pending deletion"}`))
			case 2:
				currentJSON = strings.Replace(currentJSON, "https://issuer.example.com", "https://issuer-new.example.com", 1)
				_, _ = w.Write([]byte(currentJSON))
			default:
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"kind":"Error","id":"400","reason":"invalid issuer"}`))
			}
		})
		TestServer.RouteToHandler(http.MethodDelete, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		Terraform.Source(externalAuthConfig)
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		Terraform.Source(strings.Replace(externalAuthConfig,
			"https://issuer.example.com", "https://issuer-new.example.com", 1))
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		Expect(handlerErr).ToNot(HaveOccurred())
		Expect(patchBodies).To(HaveLen(2))
		Expect(patchBodies[0]).To(Equal(patchBodies[1]))
		Expect(patchBodies[0]).To(HaveKey("issuer"))
		Expect(patchBodies[0]).ToNot(HaveKey("clients"))

		Terraform.Source(strings.Replace(externalAuthConfig,
			"https://issuer.example.com", "https://issuer-again.example.com", 1))
		output := Terraform.Apply()
		Expect(output.ExitCode).ToNot(BeZero())
		output.VerifyErrorContainsSubstring("invalid issuer")
		Expect(patchBodies).To(HaveLen(3))
		Expect(Terraform.Resource("rhcs_external_auth_provider", "example")).To(
			MatchJQ(`.attributes.issuer_url`, "https://issuer-new.example.com"),
		)
		Terraform.Source("")
		Expect(Terraform.Destroy().ExitCode).To(BeZero())
	})

	It("tolerates delete 404", func() {
		clusterJSON := externalAuthCluster(cmv1.ClusterStateReady, true, true)
		deleted := false
		TestServer.RouteToHandler(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(clusterJSON))
		})
		TestServer.RouteToHandler(http.MethodPost, externalAuthPath, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"example"}`))
		})
		TestServer.RouteToHandler(http.MethodGet, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"not found"}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":"example","issuer":{"url":"https://issuer.example.com",` +
				`"audiences":["console"]},"clients":[{"id":"console",` +
				`"component":{"name":"console","namespace":"openshift-console"}}]}`))
		})
		TestServer.RouteToHandler(http.MethodDelete, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		})
		Terraform.Source(externalAuthConfig)
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		Terraform.Source("")
		Expect(Terraform.Destroy().ExitCode).To(BeZero())
		Expect(deleted).To(BeFalse())
	})

	It("removes state when refresh returns 404", func() {
		clusterJSON := externalAuthCluster(cmv1.ClusterStateReady, true, true)
		missing := false
		TestServer.RouteToHandler(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(clusterJSON))
		})
		TestServer.RouteToHandler(http.MethodPost, externalAuthPath, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"example"}`))
		})
		TestServer.RouteToHandler(http.MethodGet, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if missing {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"not found"}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":"example","issuer":{"url":"https://issuer.example.com",` +
				`"audiences":["console"]}}`))
		})
		Terraform.Source(externalAuthConfig)
		Expect(Terraform.Apply().ExitCode).To(BeZero())
		missing = true
		Expect(Terraform.Run("refresh").ExitCode).To(BeZero())
		Expect(Terraform.State()).To(MatchJQ(`.resources | length`, 0))
	})

	It("imports an existing provider without recovering its secret", func() {
		TestServer.RouteToHandler(http.MethodGet, externalAuthPath+"/example", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"example","issuer":{"url":"https://issuer.example.com",` +
				`"audiences":["console"]},"clients":[{"id":"console",` +
				`"component":{"name":"console","namespace":"openshift-console"},` +
				`"secret":"********","type":"confidential"}]}`))
		})
		Terraform.Source(`resource "rhcs_external_auth_provider" "example" {}`)
		Expect(Terraform.Import("rhcs_external_auth_provider.example", "123,example").ExitCode).To(BeZero())
		resource := Terraform.Resource("rhcs_external_auth_provider", "example")
		Expect(resource).To(MatchJQ(`.attributes.cluster`, "123"))
		Expect(resource).To(MatchJQ(`.attributes.name`, "example"))
		Expect(resource).To(MatchJQ(`.attributes.console_client_secret`, nil))
	})

	It("reports a create conflict with import guidance", func() {
		clusterJSON := externalAuthCluster(cmv1.ClusterStateReady, true, true)
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, clusterJSON),
			),
			CombineHandlers(
				VerifyRequest(http.MethodPost, externalAuthPath),
				RespondWithJSON(http.StatusConflict, `{"message":"already exists"}`),
			),
		)
		Terraform.Source(externalAuthConfig)
		output := Terraform.Apply()
		Expect(output.ExitCode).ToNot(BeZero())
		output.VerifyErrorContainsSubstring("import it instead")
	})

	It("reports cluster lookup failures before POST", func() {
		TestServer.AppendHandlers(CombineHandlers(
			VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
			RespondWithJSON(http.StatusNotFound, `{"message":"cluster not found"}`),
		))
		Terraform.Source(externalAuthConfig)
		output := Terraform.Apply()
		Expect(output.ExitCode).ToNot(BeZero())
		output.VerifyErrorContainsSubstring("cannot retrieve cluster")
	})

	DescribeTable("rejects unsupported cluster before POST",
		func(state cmv1.ClusterState, hcpEnabled, externalAuthEnabled bool, expected string) {
			clusterJSON := externalAuthCluster(state, hcpEnabled, externalAuthEnabled)
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, clusterJSON),
				),
			)
			Terraform.Source(externalAuthConfig)
			output := Terraform.Apply()
			Expect(output.ExitCode).ToNot(BeZero())
			output.VerifyErrorContainsSubstring(expected)
		},
		Entry("not ready", cmv1.ClusterStateInstalling, true, true, "not ready"),
		Entry("not HCP", cmv1.ClusterStateReady, false, true, "Hosted Control Plane"),
		Entry("external auth disabled", cmv1.ClusterStateReady, true, false,
			"external authentication configuration is not enabled"),
	)
})
