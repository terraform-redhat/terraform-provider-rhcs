/*
Package non_cluster_e2e contains end-to-end tests that do not require a cluster.

These tests run independently without setting CLUSTER_PROFILE environment variable.

To run these tests locally:
 1. Set your OCM token: export RHCS_TOKEN=<your-token>
 2. Optional: Set OCM URL: export RHCS_URL=https://api.stage.openshift.com
 3. Build provider: make install
 4. Run tests: ginkgo tests/non_cluster_e2e

Note: The test uses placeholder AWS role ARNs. For full integration testing,
you would need actual AWS IAM roles created with proper trust policies.
*/
package non_cluster_e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Link OCM Role", ci.Day1, ci.NonClusterTest, func() {
	defer GinkgoRecover()
	var (
		profileHandler     profilehandler.ProfileHandler
		ocmRoleLinkService exec.OCMRoleLinkService
		testRoleArn        string
	)

	BeforeEach(func() {
		var err error

		// Link OCM role doesn't require a cluster profile
		// Use standalone profile handler for workspace management
		profileHandler, err = profilehandler.NewStandaloneProfileHandler("ocm-role-link")
		Expect(err).ToNot(HaveOccurred())

		ocmRoleLinkService, err = profileHandler.Services().GetOCMRoleLinkService()
		Expect(err).ToNot(HaveOccurred())

		// Generate a test role ARN
		// In a real e2e test, this should be an actual AWS IAM role ARN that exists
		testRoleName := "rhcse2e-OCM-Role-" + helper.GenerateRandomStringWithSymbols(8)
		testRoleArn = "arn:aws:iam::123456789012:role/" + testRoleName
	})

	AfterEach(func() {
		if ocmRoleLinkService != nil {
			_, err := ocmRoleLinkService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("Verify link OCM role can be created and destroyed - [id:ocm-role-link-01]", ci.Critical, func() {
		By("Create link OCM role")
		ocmRoleLinkArgs := &exec.OCMRoleLinkArgs{
			RoleArn: helper.StringPointer(testRoleArn),
		}
		_, err := ocmRoleLinkService.Apply(ocmRoleLinkArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify link OCM role output")
		ocmRoleLinkOutput, err := ocmRoleLinkService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(ocmRoleLinkOutput.RoleArn).To(Equal(testRoleArn))
		Expect(ocmRoleLinkOutput.OrganizationID).ToNot(BeEmpty())

		Logger.Infof("Successfully created link OCM role with ARN %s for organization %s",
			testRoleArn, ocmRoleLinkOutput.OrganizationID)

		By("Destroy link OCM role")
		_, err = ocmRoleLinkService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		Logger.Infof("Successfully destroyed link OCM role %s", testRoleArn)
	})

	It("Verify link OCM role validates role ARN correctly - [id:ocm-role-link-02]", ci.Medium, func() {
		By("Create link OCM role without role_arn set")
		ocmRoleLinkArgs := &exec.OCMRoleLinkArgs{}
		output, err := ocmRoleLinkService.Plan(ocmRoleLinkArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring("No value for required variable"))

		By("Create link OCM role with invalid role ARN format")
		ocmRoleLinkArgs = &exec.OCMRoleLinkArgs{
			RoleArn: helper.StringPointer("invalid-arn"),
		}
		output, err = ocmRoleLinkService.Plan(ocmRoleLinkArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring("role_arn must be a valid AWS IAM role ARN"))

		Logger.Infof("Successfully verified link OCM role validation")
	})
})
