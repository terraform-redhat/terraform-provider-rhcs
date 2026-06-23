// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package classic

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint

	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

var _ = Describe("Info data source", func() {
	It("Can read current account information", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/accounts_mgmt/v1/current_account"),
				RespondWithJSON(http.StatusOK, ocmRoleCurrentAccountResponse),
			),
		)

		Terraform.Source(`
		  data "rhcs_info" "info" {
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("rhcs_info", "info")
		Expect(resource).To(MatchJQ(`.attributes.account_id`, "1234567890"))
		Expect(resource).To(MatchJQ(`.attributes.account_name`, "Test User"))
		Expect(resource).To(MatchJQ(`.attributes.account_username`, "testuser"))
		Expect(resource).To(MatchJQ(`.attributes.account_email`, "testuser@example.com"))
		Expect(resource).To(MatchJQ(`.attributes.organization_id`, "org123"))
		Expect(resource).To(MatchJQ(`.attributes.organization_external_id`, "ext-org-456"))
		Expect(resource).To(MatchJQ(`.attributes.organization_name`, "Test Org"))
		Expect(resource).To(MatchJQ(`.attributes.ocm_aws_account_id`, "710019948333"))
	})
})
