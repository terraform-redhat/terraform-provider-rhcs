package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CMS "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("RHCS Provider Test", func() {
	Describe("Upgrade cluster tests", func() {

		var targetV string
		var clusterID string
		BeforeEach(OncePerOrdered, func() {
			By("Load the profile")
			profile := CI.LoadProfileYamlFileByENV()
			clusterID, err = CI.CreateRHCSClusterByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())

		})
		Context("Author:amalykhi-Critical-OCP-63153 @OCP-63153 @amalykhi", func() {
			It("Update a ROSA STS cluster with RHCS provider", CI.Upgrade,
				func() {

					clusterResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
					targetV, err = CMS.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
						CON.Z, clusterResp.Body().Version().AvailableUpgrades())
					Expect(err).ToNot(HaveOccurred())

					clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
					Expect(err).ToNot(HaveOccurred())

					clusterArgs := &EXE.ClusterCreationArgs{
						OpenshiftVersion: targetV,
					}

					err = clusterService.Apply(clusterArgs, true, true)

					By("Get the upgrade policy created for the addon upgrade")
					policyIDs, err := CMS.ListUpgradePolicies(CI.RHCSConnection, clusterID)
					policyID := policyIDs.Items().Get(0).ID()

					By("Wait the policy to be scheduled")
					err = H.WaitForUpgradePolicyToState(CI.RHCSConnection, clusterID, policyID, CON.Scheduled, 2)
					Expect(err).ToNot(HaveOccurred(), "Policy %s not started in 1 hour", policyID)

					By("Watch for the upgrade Started in 1 hour")
					err = H.WaitForUpgradePolicyToState(CI.RHCSConnection, clusterID, policyID, CON.Started, 60)
					Expect(err).ToNot(HaveOccurred(), "Policy %s not started in 1 hour", policyID)

					By("Watch for the upgrade finished in 2 hours")
					err = H.WaitForUpgradePolicyToState(CI.RHCSConnection, clusterID, policyID, CON.Completed, 3*60)
					Expect(err).ToNot(HaveOccurred(), "Policy %s not completed in 2 hours", policyID)
					if CON.GetEnvWithDefault(CON.WaitOperators, "false") == "true" && !profile.Private {
						// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
						timeout := 60
						timeoutMin := time.Duration(timeout)
						console, err := openshift.NewConsole(clusterID, CI.RHCSConnection)
						if err != nil {
							Logger.Warnf("Got error %s when config the openshift console. Return without waiting for operators ready", err.Error())
							return
						}
						_, err = openshift.RetryCMDRun(fmt.Sprintf("oc wait clusteroperators --all --for=condition=Progressing=false --kubeconfig %s --timeout %dm", console.KubePath, timeout), timeoutMin)
						Expect(err).ToNot(HaveOccurred())
					}
				})
		})
		Context("Author:amalykhi-Critical-OCP-63152 @OCP-63152 @amalykhi", func() {
			It("Upgrade ROSA STS cluster with RHCS provider", CI.Upgrade,
				func() {
					By("Load the profile")
					profile := CI.LoadProfileYamlFileByENV()
					var targetV string
					clusterID, err := CI.CreateRHCSClusterByProfile(token, profile)
					Expect(err).ToNot(HaveOccurred())

					clusterResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)

					targetV, err = CMS.GetVersionUpgradeTarget(clusterResp.Body().Version().RawID(),
						CON.Y, clusterResp.Body().Version().AvailableUpgrades())
					Expect(err).ToNot(HaveOccurred())

					By("Upgrade account-roles")
					_, err = CI.PrepareAccountRoles(token, clusterID, profile.UnifiedAccRolesPath, profile.Region, profile.MajorVersion, profile.ChannelGroup, CON.AccountRolesDir)
					Expect(err).ToNot(HaveOccurred())

					clusterService, err := EXE.NewClusterService(profile.ManifestsDIR)
					Expect(err).ToNot(HaveOccurred())

					clusterArgs := &EXE.ClusterCreationArgs{
						OpenshiftVersion: targetV,
					}

					err = clusterService.Apply(clusterArgs, true, true)

					By("Get the upgrade policy created for the addon upgrade")
					policyIDs, err := CMS.ListUpgradePolicies(CI.RHCSConnection, clusterID)
					policyID := policyIDs.Items().Get(0).ID()

					By("Wait the policy to be scheduled")
					err = H.WaitForUpgradePolicyToState(CI.RHCSConnection, clusterID, policyID, CON.Scheduled, 2)
					Expect(err).ToNot(HaveOccurred(), "Policy %s not started in 1 hour", policyID)

					By("Watch for the upgrade Started in 1 hour")
					err = H.WaitForUpgradePolicyToState(CI.RHCSConnection, clusterID, policyID, CON.Started, 60)
					Expect(err).ToNot(HaveOccurred(), "Policy %s not started in 1 hour", policyID)

					By("Watch for the upgrade finished in 2 hours")
					err = H.WaitForUpgradePolicyToState(CI.RHCSConnection, clusterID, policyID, CON.Completed, 3*60)
					Expect(err).ToNot(HaveOccurred(), "Policy %s not completed in 2 hours", policyID)

					if CON.GetEnvWithDefault(CON.WaitOperators, "false") == "true" && !profile.Private {
						// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
						timeout := 60
						timeoutMin := time.Duration(timeout)
						console, err := openshift.NewConsole(clusterID, CI.RHCSConnection)
						if err != nil {
							Logger.Warnf("Got error %s when config the openshift console. Return without waiting for operators ready", err.Error())
							return
						}
						_, err = openshift.RetryCMDRun(fmt.Sprintf("oc wait clusteroperators --all --for=condition=Progressing=false --kubeconfig %s --timeout %dm", console.KubePath, timeout), timeoutMin)
						Expect(err).ToNot(HaveOccurred())
					}
				})
		})
	})
})
