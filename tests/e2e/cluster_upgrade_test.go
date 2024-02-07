package e2e

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CMS "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("RHCS Provider Test", func() {
	Describe("Upgrade cluster tests", func() {

		var targetV string
		var clusterID string
		BeforeEach(OncePerOrdered, func() {
			clusterID, err = CI.PrepareRHCSClusterByProfileENV()
			Expect(err).ToNot(HaveOccurred())

		})
		Context("Author:amalykhi-Critical-OCP-63153 @OCP-63153 @amalykhi", func() {
			It("Z-stream upgrade a ROSA STS cluster with RHCS provider", CI.Upgrade,
				func() {
					if profile.VersionPattern != "z-1" {
						Skip("The test is configured only for Z-stream upgrade")
					}
					clusterResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
					targetV, err = CMS.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
						CON.Z, clusterResp.Body().Version().AvailableUpgrades())
					Expect(err).ToNot(HaveOccurred())

					clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
					Expect(err).ToNot(HaveOccurred())

					By("Validate invalid OCP version - downgrade")
					currentVersion := string(clusterResp.Body().Version().RawID())
					splittedVersion := strings.Split(currentVersion, ".")
					zStreamV, err := strconv.Atoi(splittedVersion[2])
					Expect(err).ToNot(HaveOccurred())

					downgradedVersion := fmt.Sprintf("%s.%s.%s", splittedVersion[0], splittedVersion[1], fmt.Sprint(zStreamV-1))

					imageVersionsList := CMS.EnabledVersions(CI.RHCSConnection, profile.ChannelGroup, profile.MajorVersion, true)
					versionsList := CMS.GetRawVersionList(imageVersionsList)
					if slices.Contains(versionsList, downgradedVersion) {
						clusterArgs := &EXE.ClusterCreationArgs{
							OpenshiftVersion: downgradedVersion,
						}
						err = clusterService.Apply(clusterArgs, false, false)
						Expect(err).To(HaveOccurred())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("cluster version is already above the\nrequested version"))

					}

					By("Run the cluster update")
					clusterArgs := &EXE.ClusterCreationArgs{
						OpenshiftVersion: targetV,
					}
					err = clusterService.Apply(clusterArgs, false, false)
					Expect(err).ToNot(HaveOccurred())

					By("Wait the upgrade finished")
					err = openshift.WaitClusterUpgradeFinished(CI.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred(), "Cluster upgrade %s failed with the error %v", clusterID, err)

					By("Check the cluster status and OCP version")
					clusterResp, err = CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(clusterResp.Body().State())).To(Equal(CON.Ready))
					Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

					if CON.GetEnvWithDefault(CON.WaitOperators, "false") == "true" && !profile.Private {
						// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
						timeout := 60
						err = openshift.WaitForOperatorsToBeReady(CI.RHCSConnection, clusterID, timeout)
						Expect(err).ToNot(HaveOccurred())
					}
				})
		})
		Context("Author:amalykhi-Critical-OCP-63152 @OCP-63152 @amalykhi", func() {
			It("Y-stream Upgrade ROSA STS cluster with RHCS provider", CI.Upgrade,
				func() {

					if profile.VersionPattern != "y-1" {
						Skip("The test is configured only for Y-stream upgrade")
					}

					clusterResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)

					targetV, err = CMS.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
						CON.Y, clusterResp.Body().Version().AvailableUpgrades())
					Expect(err).ToNot(HaveOccurred())
					Expect(targetV).ToNot(Equal(""))

					By("Upgrade account-roles")
					majorVersion := CI.GetMajorVersion(targetV)
					Expect(majorVersion).ToNot(Equal(""))
					_, err = CI.PrepareAccountRoles(token, clusterResp.Body().Name(), profile.UnifiedAccRolesPath, profile.Region, majorVersion, profile.ChannelGroup, CON.AccountRolesDir)
					Expect(err).ToNot(HaveOccurred())

					clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
					Expect(err).ToNot(HaveOccurred())

					By("Validate invalid OCP version field - downgrade")
					currentVersion := string(clusterResp.Body().Version().RawID())
					splittedVersion := strings.Split(currentVersion, ".")
					yStreamV, err := strconv.Atoi(splittedVersion[1])
					Expect(err).ToNot(HaveOccurred())

					downgradedVersion := fmt.Sprintf("%s.%s.%s", splittedVersion[0], fmt.Sprint(yStreamV-1), splittedVersion[2])
					imageVersionsList := CMS.EnabledVersions(CI.RHCSConnection, profile.ChannelGroup, profile.MajorVersion, true)
					versionsList := CMS.GetRawVersionList(imageVersionsList)
					if slices.Contains(versionsList, downgradedVersion) {
						clusterArgs := &EXE.ClusterCreationArgs{
							OpenshiftVersion: downgradedVersion,
						}
						err = clusterService.Apply(clusterArgs, false, false)
						Expect(err).To(HaveOccurred())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("cluster version is already above the\nrequested version"))

					}

					By("Validate  the cluster Upgrade upgrade_acknowledge field")

					clusterArgs := &EXE.ClusterCreationArgs{
						OpenshiftVersion: targetV,
					}

					err = clusterService.Apply(clusterArgs, false, false)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Missing required acknowledgements to schedule upgrade"))

					By("Apply the cluster Upgrade")

					clusterArgs = &EXE.ClusterCreationArgs{
						OpenshiftVersion:           targetV,
						UpgradeAcknowledgementsFor: majorVersion,
					}

					err = clusterService.Apply(clusterArgs, false, false)
					Expect(err).ToNot(HaveOccurred())

					By("Wait the upgrade finished")
					err = openshift.WaitClusterUpgradeFinished(CI.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)

					By("Check the cluster status and OCP version")
					clusterResp, err = CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(clusterResp.Body().State())).To(Equal(CON.Ready))
					Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

					if CON.GetEnvWithDefault(CON.WaitOperators, "false") == "true" && !profile.Private {
						// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
						timeout := 60
						err = openshift.WaitForOperatorsToBeReady(CI.RHCSConnection, clusterID, timeout)
						Expect(err).ToNot(HaveOccurred())
					}
				})
		})
	})
})
