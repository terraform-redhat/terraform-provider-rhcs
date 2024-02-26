package e2e

import (
	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("RHCS account roles Test", func() {
	Describe("RHCS account roles tests", func() {
		var err error
		var profile *CI.Profile
		var accService *EXE.AccountRoleService

		BeforeEach(func() {
			profile = CI.LoadProfileYamlFileByENV()
			Expect(err).ToNot(HaveOccurred())

			accService, err = EXE.NewAccountRoleService(CON.GetAddAccountRoleDefaultManifestDir(profile.GetClusterType()))
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			accService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("Author:amalykhi-Medium-OCP-65380 @OCP-65380 @amalykhi", func() {
			It("User can create account roles with default prefix", CI.Day2, CI.Medium, func() {
				args := &EXE.AccountRolesArgs{
					AccountRolePrefix: "",
					OpenshiftVersion:  profile.MajorVersion,
					ChannelGroup:      profile.ChannelGroup,
				}
				accRoleOutput, err := accService.Apply(args, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(accRoleOutput.AccountRolePrefix).Should(ContainSubstring(CON.DefaultAccountRolesPrefix))

			})
		})
		Context("Author:amalykhi-Critical-OCP-63316 @OCP-63316 @amalykhi", func() {
			It("User can delete account roles via account-role module", CI.Day2, CI.Critical, func() {
				args := &EXE.AccountRolesArgs{
					AccountRolePrefix: "OCP-63316",
					OpenshiftVersion:  profile.MajorVersion,
					ChannelGroup:      profile.ChannelGroup,
				}
				_, err := accService.Apply(args, true)
				Expect(err).ToNot(HaveOccurred())
				accService.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
