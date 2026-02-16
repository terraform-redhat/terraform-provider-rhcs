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
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Upgrade", func() {
	defer GinkgoRecover()

	var (
		targetV           string
		availableUpgrades []string
		clusterID         string
		profileHandler    profilehandler.ProfileHandler
		profile           profilehandler.ProfileSpec
		clusterArgs       *exec.ClusterArgs
		clusterService    exec.ClusterService
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())
		profile = profileHandler.Profile()

		clusterID, err = profileHandler.RetrieveClusterID()
		Expect(err).ToNot(HaveOccurred())

		By("Retrieve cluster args")
		clusterService, err = profileHandler.Services().GetClusterService()
		Expect(err).ToNot(HaveOccurred())
		clusterArgs, err = clusterService.ReadTFVars()
		Expect(err).ToNot(HaveOccurred())

		clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		availableUpgrades, err = cms.GetVersionUpgrades(cms.RHCSConnection, clusterResp.Body().Version().ID())
		Expect(err).ToNot(HaveOccurred())
	})

	It("on Z-stream - [id:63153] [id:72474]", ci.Upgrade,
		func() {
			if profile.GetVersion() != constants.VersionZStream {
				Skip("The test is configured only for Z-stream upgrade")
			}
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			targetV, err = cms.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
				constants.Z, availableUpgrades)
			Expect(err).ToNot(HaveOccurred())

			By("Validate invalid OCP version - downgrade")
			currentVersion := string(clusterResp.Body().Version().RawID())
			splittedVersion := strings.Split(currentVersion, ".")
			zStreamV, err := strconv.Atoi(splittedVersion[2])
			Expect(err).ToNot(HaveOccurred())

			downgradedVersion := fmt.Sprintf("%s.%s.%s", splittedVersion[0], splittedVersion[1], fmt.Sprint(zStreamV-1))

			imageVersionsList := cms.EnabledVersions(cms.RHCSConnection, profile.GetChannelGroup(), profile.GetMajorVersion(), true)
			versionsList := cms.GetRawVersionList(imageVersionsList)
			if slices.Contains(versionsList, downgradedVersion) {
				clusterArgs.OpenshiftVersion = helper.StringPointer(downgradedVersion)
				_, err = clusterService.Apply(clusterArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster version is already above the\nrequested version"))

			}

			By("Run the cluster update")
			clusterArgs.OpenshiftVersion = helper.StringPointer(targetV)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Wait the upgrade finished")
			if profile.GetClusterType().HCP {
				err = openshift.WaitHCPClusterUpgradeFinished(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)
			} else {
				err = openshift.WaitClassicClusterUpgradeFinished(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)
			}

			By("Wait for 10 minutes to be sure the version is synced in clusterdeployment")
			time.Sleep(10 * time.Minute)

			By("Check the cluster status and OCP version")
			clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusterResp.Body().State())).To(Equal(constants.Ready))
			Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

			if config.IsWaitForOperators() && !profile.IsPrivate() {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(cms.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("on Y-stream - [id:63152] [id:72475]", ci.Upgrade,
		func() {
			if profile.GetVersion() != constants.VersionYStream {
				Skip("The test is configured only for Y-stream upgrade")
			}

			By("Get cluster info")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			targetV, err = cms.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
				constants.Y, availableUpgrades)
			Expect(err).ToNot(HaveOccurred())
			Expect(targetV).ToNot(Equal(""))

			By("Upgrade account-roles")
			majorVersion := helper.GetMajorVersion(targetV)
			Expect(majorVersion).ToNot(Equal(""))
			_, err = profileHandler.Prepare().PrepareAccountRoles(token, clusterResp.Body().Name(), profile.GetUnifiedAccRolesPath(), majorVersion, profile.GetChannelGroup(), "")
			Expect(err).ToNot(HaveOccurred())

			By("Validate invalid OCP version field - downgrade")
			currentVersion := string(clusterResp.Body().Version().RawID())
			splittedVersion := strings.Split(currentVersion, ".")
			yStreamV, err := strconv.Atoi(splittedVersion[1])
			Expect(err).ToNot(HaveOccurred())

			downgradedVersion := fmt.Sprintf("%s.%s.%s", splittedVersion[0], fmt.Sprint(yStreamV-1), splittedVersion[2])
			imageVersionsList := cms.EnabledVersions(cms.RHCSConnection, profile.GetChannelGroup(), profile.GetMajorVersion(), true)
			versionsList := cms.GetRawVersionList(imageVersionsList)
			if slices.Contains(versionsList, downgradedVersion) {
				clusterArgs.OpenshiftVersion = helper.StringPointer(downgradedVersion)
				_, err = clusterService.Apply(clusterArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster version is already above the\nrequested version"))

			}

			// Blocked by OCM-7641
			// By("Validate  the cluster Upgrade upgrade_acknowledge field")
			// clusterArgs.OpenshiftVersion = helper.StringPointer(targetV)
			// _, err = clusterService.Apply(clusterArgs)
			// Expect(err).To(HaveOccurred())
			// Expect(err.Error()).To(ContainSubstring("Missing required acknowledgements to schedule upgrade"))

			By("Change the cluster channel")
			targetChannel := fmt.Sprintf("%s-%s", *clusterArgs.ChannelGroup, majorVersion)
			Logger.Infof("Setting channel for cluster '%s' to '%s'", clusterID, targetChannel)
			err = cms.ChangeClusterChannel(cms.RHCSConnection, clusterID, targetChannel)
			Expect(err).ToNot(HaveOccurred())

			By("Wait for the cluster channel to update")
			Eventually(func() []string {
				clusterDetails, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Logger.Infof("Found these available upgrades: %s", strings.Join(clusterDetails.Body().Version().AvailableUpgrades(), ","))
				return clusterDetails.Body().Version().AvailableUpgrades()
			}).WithTimeout(time.Minute * 1).Should(ContainElement(targetV))

			Logger.Infof("Starting cluster upgrade to version '%s' on channel '%s'", targetV, targetChannel)

			By("Apply the cluster Upgrade")
			clusterArgs.OpenshiftVersion = helper.StringPointer(targetV)
			clusterArgs.UpgradeAcknowledgementsFor = helper.StringPointer(majorVersion)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Wait the upgrade finished")
			if profile.GetClusterType().HCP {
				err = openshift.WaitHCPClusterUpgradeFinished(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)
			} else {
				err = openshift.WaitClassicClusterUpgradeFinished(cms.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred(), "Cluster %s failed with the error %v", clusterID, err)
			}

			By("Wait for 10 minutes to be sure the version is synced in clusterdeployment")
			time.Sleep(10 * time.Minute)

			By("Check the cluster status and OCP version")
			clusterResp, err = cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusterResp.Body().State())).To(Equal(constants.Ready))
			Expect(string(clusterResp.Body().Version().RawID())).To(Equal(targetV))

			if config.IsWaitForOperators() && !profile.IsPrivate() {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(cms.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})
})
