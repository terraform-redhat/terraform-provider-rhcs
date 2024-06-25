package e2e

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Delete cluster", func() {
	It("DestroyClusterByProfile", ci.Destroy,
		func() {
			// Destroy kubeconfig folder
			if _, err := os.Stat(constants.RHCS.KubeConfigDir); err == nil {
				os.RemoveAll(constants.RHCS.KubeConfigDir)
			}

			// Generate/build cluster by profile selected
			profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
			Expect(err).ToNot(HaveOccurred())
			err = profileHandler.DestroyRHCSClusterResources(token)
			Expect(err).ToNot(HaveOccurred())
		})
})
