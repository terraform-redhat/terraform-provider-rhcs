package cms

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"                      // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("CMS Versions Helper", func() {
	Context("GetVersionUpgrades works well", func() {
		It("version with upgrades", func() {
			version := "openshift-v4.19.0"
			By("Prepare the server")
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, fmt.Sprintf("/api/clusters_mgmt/v1/versions/%s", version)),
					RespondWithJSON(http.StatusOK, `{
					  "kind":"Version",
					  "id":"openshift-v4.19.0",
					  "href":"/api/clusters_mgmt/v1/versions/openshift-v4.19.0",
					  "raw_id":"4.19.0",
					  "enabled":true,
					  "channel_group":"stable",
					  "available_upgrades": [
					    "4.19.1",
					    "4.19.2",
					    "4.19.3"
					  ],
					  "available_channels": [
					    "eus-4.20",
					    "stable-4.19",
					    "stable-4.20"
					  ]
					}`),
				),
			)

			By("Get the available upgrades")
			versions, err := GetVersionUpgrades(connection, version)

			By("Check the available upgrades")
			Expect(err).ToNot(HaveOccurred())
			Expect(versions).ToNot(BeEmpty())
			Expect(versions).To(Equal([]string{"4.19.1", "4.19.2", "4.19.3"}))
		})

		It("version without upgrades", func() {
			version := "openshift-v4.19.0"
			By("Prepare the server")
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, fmt.Sprintf("/api/clusters_mgmt/v1/versions/%s", version)),
					RespondWithJSON(http.StatusOK, `{
					  "kind":"Version",
					  "id":"openshift-v4.19.0",
					  "href":"/api/clusters_mgmt/v1/versions/openshift-v4.19.0",
					  "raw_id":"4.19.0",
					  "enabled":true,
					  "channel_group":"stable",
					  "available_upgrades": [],
					  "available_channels": [
					    "eus-4.20",
					    "stable-4.19",
					    "stable-4.20"
					  ]
					}`),
				),
			)

			By("Get the available upgrades")
			versions, err := GetVersionUpgrades(connection, version)

			By("Check the available upgrades")
			Expect(err).ToNot(HaveOccurred())
			Expect(versions).To(BeEmpty())
		})

		It("invalid version", func() {
			version := "4.19.0"
			By("Prepare the server")
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, fmt.Sprintf("/api/clusters_mgmt/v1/versions/%s", version)),
					RespondWithJSON(http.StatusNotFound, `{
					  "kind":"Error",
					  "id":"404",
					  "href":"/api/clusters_mgmt/v1/errors/404",
					  "code":"CLUSTERS-MGMT-404",
					  "reason":"Version '4.19.0' not found",
					  "operation_id":"2fd9f62e-75b0-4d31-a66a-d253add6f31b",
					  "timestamp":"2026-03-10T16:45:13.333975472Z"
					}`),
				),
			)

			By("Get the available upgrades")
			versions, err := GetVersionUpgrades(connection, version)

			By("Check the available upgrades")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Version '4.19.0' not found"))
			Expect(versions).To(BeEmpty())
		})
	})

	Context("ChangeClusterChannel", func() {
		It("can change cluster channel", func() {
			By("Prepare the server")
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(".channel", "4.19"),
					RespondWithJSON(http.StatusOK, `{
						"kind": "Cluster",
						"id": "123",
						"href": "/api/clusters_mgmt/v1/clusters/123",
						"name": "cluster",
						"state": "ready",
						"channel": "4.19"
					}`),
				),
			)

			By("Change the channel")
			err := ChangeClusterChannel(connection, "123", "4.19")
			By("Check that it succeeded")
			Expect(err).ToNot(HaveOccurred())
		})

		It("invalid channel", func() {
			By("Prepare the server")
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(".channel", "invalid"),
					RespondWithJSON(http.StatusBadRequest, `{
					  "kind":"Error",
					  "id":"400",
					  "href":"/api/clusters_mgmt/v1/errors/400",
					  "code":"CLUSTERS-MGMT-400",
					  "reason":"Invalid channel format: 'invalid'. Channel must follow Y-stream format (e.g., stable-4.16, eus-4.16)",
					  "operation_id":"2e685ee0-1237-4767-9558-eae87246da35",
					  "timestamp":"2026-03-17T19:29:04.029369755Z"
					}`),
				),
			)
			By("Change the channel")
			err := ChangeClusterChannel(connection, "123", "invalid")
			By("Check that it failed")
			Expect(err).To(HaveOccurred())
		})

		It("invalid cluster ID", func() {
			By("Prepare the server")
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"),
					VerifyJQ(".channel", "stable-4.19"),
					RespondWithJSON(http.StatusNotFound, `{
					  "kind":"Error",
					  "id":"404",
					  "href":"/api/clusters_mgmt/v1/errors/404",
					  "code":"CLUSTERS-MGMT-404",
					  "reason":"Cluster '123' not found",
					  "operation_id":"d8c32fec-345c-4ac6-b8f5-b092cd573229",
					  "timestamp":"2026-03-17T19:33:45.173193619Z"
					}`),
				),
			)
			By("Change the channel")
			err := ChangeClusterChannel(connection, "123", "stable-4.19")
			By("Check that it failed")
			Expect(err).To(HaveOccurred())
		})
	})
})
