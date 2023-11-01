package e2e

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
)

var _ = Describe("TF Test", func() {
	Describe("Create cluster test", func() {
		It("DestroyClusterByProfile", CI.Destroy,
			func() {
				// Destroy kubeconfig folder
				if _, err := os.Stat(CON.RHCS.KubeConfigDir); err == nil {
					os.RemoveAll(CON.RHCS.KubeConfigDir)
				}

				// Generate/build cluster by profile selected
				profile := CI.LoadProfileYamlFileByENV()
				err := CI.DestroyRHCSClusterByProfile(token, profile)

				Expect(err).ToNot(HaveOccurred())
			})
	})
})
