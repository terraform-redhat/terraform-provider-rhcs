package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	helper "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("RHCS Provider Negative Test", func() {
	var originalCreationArgs *exe.ClusterCreationArgs
	var creationArgs *exe.ClusterCreationArgs
	var clusterService *exe.ClusterService
	var err error

	Describe("Cluster admin Negative tests", ci.NonHCPCluster, Ordered, func() {
		BeforeAll(func() {
			if !profile.AdminEnabled {
				Skip("The tests configured for cluster admin only")
			}
			originalCreationArgs, err = ci.GenerateClusterCreationArgsByProfile(token, profile)
			if err != nil {
				defer ci.DestroyRHCSClusterByProfile(token, profile)
			}
			Expect(err).ToNot(HaveOccurred())
			tfExecHelper, err := ci.GetTerraformExecHelperForProfile(profile)
			if err != nil {
				defer ci.DestroyRHCSClusterByProfile(token, profile)
			}
			Expect(err).ToNot(HaveOccurred())
			clusterService, err = tfExecHelper.GetClusterService()
			if err != nil {
				defer ci.DestroyRHCSClusterByProfile(token, profile)
			}
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			err := ci.DestroyRHCSClusterByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
		})

		BeforeEach(OncePerOrdered, func() {
			// Restore cluster args
			creationArgs = originalCreationArgs
		})
		Context("Author:amalykhi-Medium-OCP-65961 @OCP-65961 @amalykhi", func() {
			It("Cluster admin during deployment - validate user name policy", ci.Day1Negative,
				func() {
					By("Edit cluster admin user name to not valid")
					creationArgs.AdminCredentials["username"] = "one:two"
					_, err = clusterService.Apply(creationArgs)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.username username may not contain the characters:\n'/:%'"))

					By("Edit cluster admin user name to empty")
					creationArgs.AdminCredentials["username"] = ""
					_, err = clusterService.Apply(creationArgs)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Attribute 'username' is mandatory"))
				})
		})
		Context("Author:amalykhi-Medium-OCP-65963 @OCP-65963 @amalykhi", func() {
			It("Cluster admin during deployment - validate password policy", ci.Day1Negative, func() {
				By("Edit cluster admin password  to the short one")
				creationArgs.AdminCredentials["password"] = helper.GenerateRandomStringWithSymbols(13)
				_, err = clusterService.Apply(creationArgs)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password string length must be at least 14"))
				By("Edit cluster admin password to empty")
				creationArgs.AdminCredentials["password"] = ""
				_, err = clusterService.Apply(creationArgs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password should use ASCII-standard"))

				By("Edit cluster admin password that lacks a capital letter")
				creationArgs.AdminCredentials["password"] = strings.ToLower(helper.GenerateRandomStringWithSymbols(14))
				_, err = clusterService.Apply(creationArgs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password must contain uppercase\ncharacters"))

				By("Edit cluster admin password that lacks symbol but has digits")
				creationArgs.AdminCredentials["password"] = "QwertyPasswordNoDigitsSymbols"
				_, err = clusterService.Apply(creationArgs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password must contain numbers or\nsymbols"))

				By("Edit cluster admin password that includes Non English chars")
				creationArgs.AdminCredentials["password"] = "Qwert12345345@ש"
				_, err = clusterService.Apply(creationArgs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Attribute admin_credentials.password password should use ASCII-standard\ncharacters only"))

			})
		})
	})
})
