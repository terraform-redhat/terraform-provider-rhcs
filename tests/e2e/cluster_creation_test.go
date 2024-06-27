package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Create cluster", func() {
	It("CreateClusterByProfile", ci.Day1Prepare,
		func() {

			// Generate/build cluster by profile selected
			profile := ci.LoadProfileYamlFileByENV()
			clusterID, err := ci.CreateRHCSClusterByProfile(token, profile)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).ToNot(BeEmpty())
			//TODO: implement waiter for  the private cluster once bastion is implemented
			if constants.GetEnvWithDefault(constants.WaitOperators, "false") == "true" && !profile.Private {
				// WaitClusterOperatorsToReadyStatus will wait for cluster operators ready
				timeout := 60
				err = openshift.WaitForOperatorsToBeReady(ci.RHCSConnection, clusterID, timeout)
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("Cluster can be recreated if it was not deleted from tf - [id:66071]", ci.Day3, ci.Medium,
		func() {

			profile := ci.LoadProfileYamlFileByENV()
			originalClusterID := clusterID

			By("Delete cluster via OCM")
			resp, err := cms.DeleteCluster(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Status()).To(Equal(constants.HTTPNoContent))

			By("Wait for the cluster deleted from OCM")
			err = cms.WaitClusterDeleted(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Create cluster again")
			clusterService, err := exec.NewClusterService(profile.GetClusterManifestsDir())
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
})
