package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
)

var _ = Describe("TF Test", func() {
	Describe("Create cluster test", func() {
		It("DestroyClusterByProfile", CI.Destroy,
			func() {

				// Generate/build cluster by profile selected
				profile := CI.LoadProfileYamlFileByENV()
				err := CI.DestroyRHCSClusterByProfile(token, profile)
				Expect(err).ToNot(HaveOccurred())
			})
	})
})
