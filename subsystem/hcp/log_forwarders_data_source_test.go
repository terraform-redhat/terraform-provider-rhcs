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

var _ = Describe("Log forwarders data source", func() {
	It("fails if cluster ID is empty", func() {
		Terraform.Source(`
		  data "rhcs_log_forwarders" "all" {
		    cluster = ""
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		runOutput.VerifyErrorContainsSubstring("cluster ID may not be empty/blank string")
	})

	It("Can list log forwarders for a cluster", func() {
		listResponse := `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "log-fwd-1",
				      "s3": {
				        "bucket_name": "my-logs",
				        "bucket_prefix": "rosa/"
				      },
				      "applications": ["audit", "infrastructure"]
				    },
				    {
				      "id": "log-fwd-2",
				      "cloudwatch": {
				        "log_group_name": "/aws/rosa/logs",
				        "log_distribution_role_arn": "arn:aws:iam::123456789012:role/log-forwarder"
				      },
				      "applications": ["audit"]
				    }
				  ]
				}`
		listHandler := CombineHandlers(
			VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/456/control_plane/log_forwarders"),
			RespondWithJSON(http.StatusOK, listResponse),
		)
		// Data source reads run during plan and apply.
		TestServer.AppendHandlers(listHandler, listHandler)

		Terraform.Source(`
		  data "rhcs_log_forwarders" "all" {
		    cluster = "456"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("rhcs_log_forwarders", "all")
		Expect(resource).To(MatchJQ(`.attributes.cluster`, "456"))
		Expect(resource).To(MatchJQ(`.attributes.items | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "log-fwd-1"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].cluster_id`, "456"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].s3.bucket_name`, "my-logs"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].s3.bucket_prefix`, "rosa/"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].applications[0]`, "audit"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "log-fwd-2"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].cloudwatch.log_group_name`, "/aws/rosa/logs"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].cloudwatch.log_distribution_role_arn`, "arn:aws:iam::123456789012:role/log-forwarder"))
	})
})
