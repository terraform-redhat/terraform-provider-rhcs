package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Create cluster", func() {
	It("CreateClusterByProfile", CI.Day1Prepare,
		func() {

			// Generate/build cluster by profile selected
			profile := CI.LoadProfileYamlFileByENV()
			clusterID, err := CI.CreateRHCSClusterByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())
			//TODO: implement waiter for  the private cluster once bastion is implemented
			if CON.GetEnvWithDefault(CON.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(CI.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("Cluster can be recreated if it was not deleted from tf - [id:66071]", CI.Day3, CI.Medium,
		func() {

			profile := CI.LoadProfileYamlFileByENV()
			originalClusterID := clusterID

			By("Delete cluster via OCM")
			resp, err := cms.DeleteCluster(CI.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Status()).To(Equal(CON.HTTPNoContent))

			By("Wait for the cluster deleted from OCM")
			err = cms.WaitClusterDeleted(CI.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Create cluster again")
			clusterArgs := &exec.ClusterCreationArgs{}
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
			Expect(err).ToNot(HaveOccurred())
			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).ToNot(HaveOccurred())
			clusterOutput, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterOutput.ClusterID).ToNot(BeEmpty())
			Expect(clusterOutput.ClusterID).ToNot(Equal(originalClusterID))
		})
})
