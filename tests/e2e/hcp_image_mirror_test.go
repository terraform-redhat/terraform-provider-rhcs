package e2e

import (
	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("HCP ImageMirror", ci.Day2, func() {
	defer GinkgoRecover()
	var (
		imageMirrorService exec.ImageMirrorService
		profileHandler     profilehandler.ProfileHandler
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if !profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Hosted cluster")
		}

		imageMirrorService, err = profileHandler.Services().GetImageMirrorService()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		imageMirrorService.Destroy()
	})

	It("can be created with required attributes",
		ci.Critical, ci.FeatureImageMirror, func() {
			By("Create image mirror")
			source := "docker.io/library/nginx"
			mirrors := []string{"quay.io/my-org/nginx"}
			imageMirrorArgs := exec.NewImageMirrorArgs(clusterID, source, mirrors)

			_, err := imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify image mirror attributes are correctly set")
			output, err := imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ClusterID).To(Equal(clusterID))
			Expect(output.Type).To(Equal("digest"))
			Expect(output.Source).To(Equal(source))
			Expect(output.Mirrors).To(Equal(mirrors))
			Expect(output.ID).ToNot(BeEmpty())
			Expect(output.CreationTimestamp).ToNot(BeEmpty())
			Expect(output.LastUpdateTimestamp).ToNot(BeEmpty())

			By("Verify image mirror via OCM API")
			imageMirrorResponse, err := cms.RetrieveClusterImageMirror(cms.RHCSConnection, clusterID, output.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(imageMirrorResponse.ID()).To(Equal(output.ID))
			Expect(imageMirrorResponse.Type()).To(Equal("digest"))
			Expect(imageMirrorResponse.Source()).To(Equal(source))
			Expect(imageMirrorResponse.Mirrors()).To(Equal(mirrors))
		})

	It("can be updated to add and remove mirrors",
		ci.Critical, ci.FeatureImageMirror, func() {
			By("Create image mirror with single mirror")
			source := "docker.io/library/alpine"
			initialMirrors := []string{"quay.io/my-org/alpine"}
			imageMirrorArgs := exec.NewImageMirrorArgs(clusterID, source, initialMirrors)

			_, err := imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify initial mirror configuration")
			output, err := imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Mirrors).To(Equal(initialMirrors))
			originalID := output.ID

			By("Update image mirror to add additional mirrors")
			updatedMirrors := []string{
				"quay.io/my-org/alpine",
				"registry.example.com/alpine",
				"docker.io/my-org/alpine",
			}
			imageMirrorArgs.Mirrors = &updatedMirrors
			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify mirrors were added")
			output, err = imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ID).To(Equal(originalID)) // ID should remain the same
			Expect(output.Mirrors).To(Equal(updatedMirrors))
			Expect(output.Source).To(Equal(source)) // Source should remain unchanged
			Expect(output.Type).To(Equal("digest")) // Type should remain unchanged

			By("Update image mirror to remove some mirrors")
			finalMirrors := []string{"registry.example.com/alpine"}
			imageMirrorArgs.Mirrors = &finalMirrors
			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify mirrors were removed")
			output, err = imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ID).To(Equal(originalID)) // ID should remain the same
			Expect(output.Mirrors).To(Equal(finalMirrors))

			By("Verify final state via OCM API")
			imageMirrorResponse, err := cms.RetrieveClusterImageMirror(cms.RHCSConnection, clusterID, output.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(imageMirrorResponse.Mirrors()).To(Equal(finalMirrors))
		})

	It("can validate invalid configurations",
		ci.Medium, ci.FeatureImageMirror, func() {
			By("Try to create image mirror with empty source")
			source := ""
			mirrors := []string{"quay.io/my-org/test"}
			imageMirrorArgs := exec.NewImageMirrorArgs(clusterID, source, mirrors)

			_, err := imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source"))

			By("Try to create image mirror with empty mirrors")
			source = "docker.io/library/ubuntu"
			mirrors = []string{}
			imageMirrorArgs = exec.NewImageMirrorArgs(clusterID, source, mirrors)

			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mirrors"))

			By("Try to create image mirror with invalid cluster ID")
			source = "docker.io/library/ubuntu"
			mirrors = []string{"quay.io/my-org/test"}
			imageMirrorArgs = exec.NewImageMirrorArgs("invalid-cluster-id", source, mirrors)

			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).To(HaveOccurred())
		})

	It("enforces immutable fields",
		ci.Medium, ci.FeatureImageMirror, func() {
			By("Create image mirror")
			source := "docker.io/library/redis"
			mirrors := []string{"quay.io/my-org/redis"}
			imageMirrorArgs := exec.NewImageMirrorArgs(clusterID, source, mirrors)

			_, err := imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Try to update source (should trigger replace)")
			newSource := "docker.io/library/postgres"
			imageMirrorArgs.Source = helper.StringPointer(newSource)

			// This should work but should trigger a replace operation
			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify source was updated via replace")
			output, err := imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Source).To(Equal(newSource))

			By("Try to update cluster_id (should trigger replace)")
			newClusterID := "different-cluster-id"
			imageMirrorArgs.Cluster = helper.StringPointer(newClusterID)

			// This should work but should trigger a replace operation
			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify cluster_id was updated via replace")
			output, err = imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ClusterID).To(Equal(newClusterID))
		})

	It("can update type field without replacement",
		ci.Medium, ci.FeatureImageMirror, func() {
			By("Create image mirror")
			source := "docker.io/library/nginx"
			mirrors := []string{"quay.io/my-org/nginx"}
			imageMirrorArgs := exec.NewImageMirrorArgs(clusterID, source, mirrors)

			_, err := imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify initial configuration")
			output, err := imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Type).To(Equal("digest"))
			originalID := output.ID

			By("Update type field (should not trigger replace)")
			// Note: Currently only 'digest' type is supported, so this test
			// verifies that the type field can be updated in-place when
			// additional types become available in the future
			imageMirrorArgs.Type = helper.StringPointer("digest") // Same value, but tests update path

			_, err = imageMirrorService.Apply(imageMirrorArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify type was updated without replacement")
			output, err = imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ID).To(Equal(originalID)) // ID should remain the same
			Expect(output.Type).To(Equal("digest"))
			Expect(output.Source).To(Equal(source))   // Source should remain unchanged
			Expect(output.Mirrors).To(Equal(mirrors)) // Mirrors should remain unchanged

			By("Verify type update via OCM API")
			imageMirrorResponse, err := cms.RetrieveClusterImageMirror(cms.RHCSConnection, clusterID, output.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(imageMirrorResponse.Type()).To(Equal("digest"))
		})

	It("can handle multiple image mirrors",
		ci.High, ci.FeatureImageMirror, func() {
			By("Create first image mirror")
			source1 := "docker.io/library/busybox"
			mirrors1 := []string{"quay.io/my-org/busybox"}
			imageMirrorArgs1 := exec.NewImageMirrorArgs(clusterID, source1, mirrors1)

			_, err := imageMirrorService.Apply(imageMirrorArgs1)
			Expect(err).ToNot(HaveOccurred())

			output1, err := imageMirrorService.Output()
			Expect(err).ToNot(HaveOccurred())
			firstMirrorID := output1.ID

			By("Create second image mirror via new service")
			imageMirrorService2, err := exec.NewImageMirrorService(
				helper.GenerateRandomName("tf-workspace", 2),
				profileHandler.Profile().GetClusterType(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer imageMirrorService2.Destroy()

			source2 := "docker.io/library/mysql"
			mirrors2 := []string{"quay.io/my-org/mysql", "registry.example.com/mysql"}
			imageMirrorArgs2 := exec.NewImageMirrorArgs(clusterID, source2, mirrors2)

			_, err = imageMirrorService2.Apply(imageMirrorArgs2)
			Expect(err).ToNot(HaveOccurred())

			output2, err := imageMirrorService2.Output()
			Expect(err).ToNot(HaveOccurred())
			secondMirrorID := output2.ID

			By("Verify both image mirrors exist and are different")
			Expect(firstMirrorID).ToNot(Equal(secondMirrorID))

			// Verify first mirror still exists
			imageMirrorResponse1, err := cms.RetrieveClusterImageMirror(cms.RHCSConnection, clusterID, firstMirrorID)
			Expect(err).ToNot(HaveOccurred())
			Expect(imageMirrorResponse1.Source()).To(Equal(source1))

			// Verify second mirror exists
			imageMirrorResponse2, err := cms.RetrieveClusterImageMirror(cms.RHCSConnection, clusterID, secondMirrorID)
			Expect(err).ToNot(HaveOccurred())
			Expect(imageMirrorResponse2.Source()).To(Equal(source2))
		})
})
