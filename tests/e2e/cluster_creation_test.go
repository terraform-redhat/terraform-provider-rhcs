package e2e

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Create cluster", func() {
	It("CreateClusterByProfile", ci.Day1Prepare,
		func() {
			// Generate/build cluster by profile selected
			profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
			Expect(err).ToNot(HaveOccurred())
			clusterID, err := profileHandler.CreateRHCSClusterByProfile(token)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())
			//TODO: implement waiter for  the private cluster once bastion is implemented
			if config.IsWaitForOperators() && !profileHandler.Profile().IsPrivate() {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(cms.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("Cluster can be recreated if it was not deleted from tf - [id:66071]", ci.Day3, ci.Medium,
		func() {
			profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
			Expect(err).ToNot(HaveOccurred())
			originalClusterID := clusterID

			By("Delete cluster via OCM")
			resp, err := cms.DeleteCluster(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Status()).To(Equal(constants.HTTPNoContent))

			By("Wait for the cluster deleted from OCM")
			err = cms.WaitClusterDeleted(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Create cluster again")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			clusterArgs, err := clusterService.ReadTFVars()
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
			clusterOutput, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterOutput.ClusterID).ToNot(BeEmpty())
			Expect(clusterOutput.ClusterID).ToNot(Equal(originalClusterID))
		})

	It("cluster waiter should fail on error/uninstalling cluster - [id:74472]", ci.Day1Supplemental, ci.Medium,
		func() {
			By("Retrieve random profile")
			profileHandler, err := profilehandler.NewRandomProfileHandler()
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve cluster with error/uninstalling status")
			params := map[string]interface{}{
				"search": "status.state='error' or status.state='uninstalling'",
				"size":   -1,
			}
			resp, err := cms.ListClusters(cms.RHCSConnection, params)
			Expect(err).ToNot(HaveOccurred())
			clusters := resp.Items().Slice()

			if len(clusters) <= 0 {
				Skip("No cluster in 'error' or 'uninstalling' state to launch this case")
			}
			cluster := clusters[0]

			By("Wait for cluster")
			cwService, err := profileHandler.Services().GetClusterWaiterService()
			Expect(err).ToNot(HaveOccurred())
			clusterWaiterArgs := exec.ClusterWaiterArgs{
				Cluster:      helper.StringPointer(cluster.ID()),
				TimeoutInMin: helper.IntPointer(60),
			}
			_, err = cwService.Apply(&clusterWaiterArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(`Waiting for cluster creation finished with error`))
		})

	It("Cluster can be created within internal cluster waiter - [id:74473]", ci.Day1Supplemental, ci.Medium,
		func() {
			By("Retrieve random profile")
			profileHandler, err := profilehandler.NewRandomProfileHandler()
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			defer func() {
				By("Clean resources")
				profileHandler.DestroyRHCSClusterResources(token)
			}()
			clusterArgs, err := profileHandler.GenerateClusterCreationArgs(token)
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.DisableClusterWaiter = helper.BoolPointer(true)

			By("Create cluster")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())
		})

	It("Cluster can be created without cluster waiter - [id:74748]", ci.Day1Supplemental, ci.Medium,
		func() {
			By("Retrieve random profile")
			profileHandler, err := profilehandler.NewRandomProfileHandler()
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			defer func() {
				By("Clean resources")
				profileHandler.DestroyRHCSClusterResources(token)
			}()
			clusterArgs, err := profileHandler.GenerateClusterCreationArgs(token)
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.WaitForCluster = helper.BoolPointer(false)

			By("Create cluster")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify cluster configuration")
			clusterOut, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterID := clusterOut.ClusterID
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterResp.Body().State()). // we did not wait for the cluster to be ready
								To(
					BeElementOf(
						[]v1.ClusterState{
							v1.ClusterStateInstalling,
							v1.ClusterStateWaiting,
							v1.ClusterStateValidating,
						}),
				)
		})

	It("HCP Cluster can be destroyed without waiting - [id:72500]", ci.Day3, ci.High,
		func() {
			profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve creation args")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			clusterArgs, err := clusterService.ReadTFVars()
			Expect(err).ToNot(HaveOccurred())
			clusterArgs.DisableWaitingInDestroy = helper.BoolPointer(true)

			By("Apply changes")
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Destroy cluster")
			_, err = clusterService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			// Give some time for OCM to update
			time.Sleep(5 * time.Second)

			By("Verify cluster is in uninstalling state")
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterResp.Body().State()).To(Equal(v1.ClusterStateUninstalling))
		})
})
