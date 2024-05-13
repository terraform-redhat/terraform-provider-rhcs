package e2e

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Upgrade", func() {
	defer GinkgoRecover()

	var (
		targetV   string
		clusterID string
		profile   *ci.Profile
	)

	BeforeEach(OncePerOrdered, func() {
		profile = ci.LoadProfileYamlFileByENV()

		var err error
		clusterID, err = ci.PrepareRHCSClusterByProfileENV()
		Expect(err).ToNot(HaveOccurred())

	})

	It("ROSA STS cluster on Z-stream - [id:63153]", ci.Upgrade, ci.NonHCPCluster,
		func() {
			if profile.VersionPattern != "z-1" {
				Skip("The test is configured only for Z-stream upgrade")
			}
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			targetV, err = cms.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
				constants.Z, clusterResp.Body().Version().AvailableUpgrades())
			Expect(err).ToNot(HaveOccurred())

			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())

			By("Validate invalid OCP version - downgrade")
			currentVersion := string(clusterResp.Body().Version().RawID())
			splittedVersion := strings.Split(currentVersion, ".")
			zStreamV, err := strconv.Atoi(splittedVersion[2])
			Expect(err).ToNot(HaveOccurred())

			downgradedVersion := fmt.Sprintf("%s.%s.%s", splittedVersion[0], splittedVersion[1], fmt.Sprint(zStreamV-1))

			imageVersionsList := cms.EnabledVersions(ci.RHCSConnection, profile.ChannelGroup, profile.MajorVersion, true)
			versionsList := cms.GetRawVersionList(imageVersionsList)
			if slices.Contains(versionsList, downgradedVersion) {
				clusterArgs := &exec.ClusterCreationArgs{
					OpenshiftVersion: downgradedVersion,
				}
				err = clusterService.Apply(clusterArgs, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster version is already above the\nrequested version"))

			}

			By("Run the cluster update")
			clusterArgs := &exec.ClusterCreationArgs{
				OpenshiftVersion: targetV,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).ToNot(HaveOccurred())

			By("Wait the upgrade finished")
			err = openshift.WaitClassicClusterUpgradeFinished(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred(), "Cluster upgrade %s failed with the error %v", clusterID, err)

			By("Wait for 10 minutes to be sure the version is synced in clusterdeployment")
			time.Sleep(10 * time.Minute)

			By("Check the cluster status and OCP version")
			clusterResp, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusterResp.Body().State())).To(Equal(constants.Ready))
			Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

			if constants.GetEnvWithDefault(constants.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(ci.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("ROSA STS cluster on Y-stream - [id:63152]", ci.Upgrade, ci.NonHCPCluster,
		func() {

			if profile.VersionPattern != "y-1" {
				Skip("The test is configured only for Y-stream upgrade")
			}

			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			targetV, err = cms.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
				constants.Y, clusterResp.Body().Version().AvailableUpgrades())
			Expect(err).ToNot(HaveOccurred())
			Expect(targetV).ToNot(Equal(""))

			By("Upgrade account-roles")
			majorVersion := ci.GetMajorVersion(targetV)
			Expect(majorVersion).ToNot(Equal(""))
			_, err = ci.PrepareAccountRoles(token, clusterResp.Body().Name(), profile.UnifiedAccRolesPath, profile.Region, majorVersion, profile.ChannelGroup, profile.GetClusterType(), "")
			Expect(err).ToNot(HaveOccurred())

			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())

			By("Validate invalid OCP version field - downgrade")
			currentVersion := string(clusterResp.Body().Version().RawID())
			splittedVersion := strings.Split(currentVersion, ".")
			yStreamV, err := strconv.Atoi(splittedVersion[1])
			Expect(err).ToNot(HaveOccurred())

			downgradedVersion := fmt.Sprintf("%s.%s.%s", splittedVersion[0], fmt.Sprint(yStreamV-1), splittedVersion[2])
			imageVersionsList := cms.EnabledVersions(ci.RHCSConnection, profile.ChannelGroup, profile.MajorVersion, true)
			versionsList := cms.GetRawVersionList(imageVersionsList)
			if slices.Contains(versionsList, downgradedVersion) {
				clusterArgs := &exec.ClusterCreationArgs{
					OpenshiftVersion: downgradedVersion,
				}
				err = clusterService.Apply(clusterArgs, false, false)
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster version is already above the\nrequested version"))

			}

			By("Validate  the cluster Upgrade upgrade_acknowledge field")

			clusterArgs := &exec.ClusterCreationArgs{
				OpenshiftVersion: targetV,
			}

			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Missing required acknowledgements to schedule upgrade"))

			By("Apply the cluster Upgrade")

			clusterArgs = &exec.ClusterCreationArgs{
				OpenshiftVersion:           targetV,
				UpgradeAcknowledgementsFor: majorVersion,
			}

			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).ToNot(HaveOccurred())

			By("Wait the upgrade finished")
			err = openshift.WaitClassicClusterUpgradeFinished(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)

			By("Wait for 10 minutes to be sure the version is synced in clusterdeployment")
			time.Sleep(10 * time.Minute)

			By("Check the cluster status and OCP version")
			clusterResp, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusterResp.Body().State())).To(Equal(constants.Ready))
			Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

			if constants.GetEnvWithDefault(constants.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(ci.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("ROSA HCP cluster on Z-stream - [id:72474]", ci.Upgrade, ci.NonHCPCluster,
		func() {
			if profile.VersionPattern != "z-1" {
				Skip("The test is configured only for Z-stream upgrade")
			}

			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve cluster information and upgrade version")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			targetV, err = cms.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
				constants.Z, clusterResp.Body().Version().AvailableUpgrades())
			Expect(err).ToNot(HaveOccurred())
			Expect(targetV).ToNot(BeEmpty())

			Logger.Infof("Gonna upgrade to version %s", targetV)

			By("Run the cluster update")
			clusterArgs := &exec.ClusterCreationArgs{
				OpenshiftVersion: targetV,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).ToNot(HaveOccurred())

			By("Wait the upgrade finished")
			err = openshift.WaitHCPClusterUpgradeFinished(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred(), "Cluster upgrade %s failed with the error %v", clusterID, err)

			By("Check the cluster status and OCP version")
			clusterResp, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusterResp.Body().State())).To(Equal(constants.Ready))
			Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

			if constants.GetEnvWithDefault(constants.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(ci.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("ROSA HCP cluster on Y-stream - [id:72475]", ci.Upgrade, ci.NonHCPCluster,
		func() {
			if profile.VersionPattern != "y-1" {
				Skip("The test is configured only for Y-stream upgrade")
			}

			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve cluster information and upgrade version")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			targetV, err = cms.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
				constants.Y, clusterResp.Body().Version().AvailableUpgrades())
			Expect(err).ToNot(HaveOccurred())
			Expect(targetV).ToNot(BeEmpty())
			majorVersion := ci.GetMajorVersion(targetV)
			Expect(majorVersion).ToNot(BeEmpty())

			Logger.Infof("Gonna upgrade to version %s", targetV)

			// Blocked by OCM-7641
			// By("Validate the cluster Upgrade upgrade_acknowledge field")
			// clusterArgs := &exec.ClusterCreationArgs{
			// 	OpenshiftVersion: targetV,
			// }
			// err = clusterService.Apply(clusterArgs, false, false)
			// Expect(err).To(HaveOccurred())
			// Expect(err.Error()).To(ContainSubstring("Missing required acknowledgements to schedule upgrade"))

			By("Apply the cluster Upgrade")
			clusterArgs := &exec.ClusterCreationArgs{
				OpenshiftVersion:           targetV,
				UpgradeAcknowledgementsFor: majorVersion,
			}
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).ToNot(HaveOccurred())

			By("Wait the upgrade finished")
			err = openshift.WaitHCPClusterUpgradeFinished(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)

			By("Check the cluster status and OCP version")
			clusterResp, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusterResp.Body().State())).To(Equal(constants.Ready))
			Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

			if constants.GetEnvWithDefault(constants.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(ci.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})
})
