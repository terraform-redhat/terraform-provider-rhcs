package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("RHCS Provider Test", func() {
	Describe("Create cluster test", func() {
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
	})
})
