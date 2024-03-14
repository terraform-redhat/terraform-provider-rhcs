package provider

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Trusted IP Addresses data source", func() {
	It("Can list trusted ip addresses", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/trusted_ip_addresses"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 2,
				  "items": [
				    {
				      "enabled": true,
					  "id": "1.1.1.1"
				    },
				    {
				      "enabled": false,
					  "id": "2.2.2.2"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "rhcs_trusted_ip_addresses" "my_trusted_ip_addresses" {
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("rhcs_trusted_ip_addresses", "my_trusted_ip_addresses")
		Expect(resource).To(MatchJQ(".attributes.items[0].id", "1.1.1.1"))
		Expect(resource).To(MatchJQ(".attributes.items[0].enabled", true))
		Expect(resource).To(MatchJQ(".attributes.items[1].id", "2.2.2.2"))
		Expect(resource).To(MatchJQ(".attributes.items[1].enabled", false))
	})
})
