package e2e

import (
	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("RHCS account roles Test", func() {
	Describe("RHCS account roles tests", func() {
		var err error
		var accService *exec.AccountRoleService

		BeforeEach(func() {
			tfExecHelper, err := exec.NewTerraformExecHelperWithWorkspaceName(profile.GetClusterType(), ci.GenerateNewTerraformWorkspaceFromProfile(profile))
			Expect(err).ToNot(HaveOccurred())
			accService, err = tfExecHelper.GetAccountRoleService()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			accService.Destroy(true)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("Author:amalykhi-Medium-OCP-65380 @OCP-65380 @amalykhi", func() {
			It("User can create account roles with default prefix", ci.Day2, ci.Medium, func() {
				args := &exec.AccountRolesArgs{
					AccountRolePrefix: "",
					OpenshiftVersion:  profile.MajorVersion,
					ChannelGroup:      profile.ChannelGroup,
				}
				_, err := accService.Apply(args, true)
				Expect(err).ToNot(HaveOccurred())
				accRoleOutput, err := accService.Output()
				Expect(accRoleOutput.AccountRolePrefix).Should(ContainSubstring(constants.DefaultAccountRolesPrefix))

			})
		})
		Context("Author:amalykhi-Critical-OCP-63316 @OCP-63316 @amalykhi", func() {
			It("User can delete account roles via account-role module", ci.Day2, ci.Critical, func() {
				args := &exec.AccountRolesArgs{
					AccountRolePrefix: "OCP-63316",
					OpenshiftVersion:  profile.MajorVersion,
					ChannelGroup:      profile.ChannelGroup,
				}
				_, err := accService.Apply(args, true)
				Expect(err).ToNot(HaveOccurred())
				accService.Destroy(true)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
