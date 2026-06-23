// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hcp

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("Image mirrors data source", func() {
	It("Can list image mirrors for a cluster", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/image_mirrors"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 1,
				  "items": [
				    {
				      "id": "mirror-1",
				      "type": "digest",
				      "source": "registry.example.com/team",
				      "mirrors": ["mirror.corp.com/team"],
				      "creation_timestamp": "2024-01-01T00:00:00Z",
				      "last_update_timestamp": "2024-01-02T00:00:00Z"
				    }
				  ]
				}`),
			),
		)

		Terraform.Source(`
		  data "rhcs_image_mirrors" "cluster_mirrors" {
		    cluster_id = "123"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("rhcs_image_mirrors", "cluster_mirrors")
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(`.attributes.image_mirrors | length`, 1))
		Expect(resource).To(MatchJQ(`.attributes.image_mirrors[0].id`, "mirror-1"))
		Expect(resource).To(MatchJQ(`.attributes.image_mirrors[0].type`, "digest"))
		Expect(resource).To(MatchJQ(`.attributes.image_mirrors[0].source`, "registry.example.com/team"))
		Expect(resource).To(MatchJQ(`.attributes.image_mirrors[0].mirrors[0]`, "mirror.corp.com/team"))
	})
})
