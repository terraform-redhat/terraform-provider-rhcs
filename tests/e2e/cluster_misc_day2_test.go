package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Cluster miscellaneous", func() {
	defer GinkgoRecover()

	var (
		clusterService           exec.ClusterService
		err                      error
		profile                  *ci.Profile
		originalCustomProperties map[string]string
		clusterArgs              *exec.ClusterArgs
	)

	BeforeEach(func() {
		// Load profile from YAML file based on environment
		By("Load profile")
		profile = ci.LoadProfileYamlFileByENV()

		// Initialize the cluster service
		By("Create cluster service")
		clusterService, err = exec.NewClusterService(profile.GetClusterManifestsDir())
		Expect(err).ShouldNot(HaveOccurred())

		// Read terraform.tfvars file and get its content as a map
		By("Retrieve current properties")
		clusterArgs, err = clusterService.ReadTFVars()
		Expect(err).ShouldNot(HaveOccurred())
		originalCustomProperties = *clusterArgs.CustomProperties
	})

	AfterEach(func() {
		By("Recover cluster properties")
		clusterArgs.CustomProperties = helper.StringMapPointer(originalCustomProperties)

		// Restore cluster state
		_, err = clusterService.Apply(clusterArgs)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should validate custom property operations on cluster - [id:64907]",
		ci.Day2, ci.Medium, ci.FeatureClusterMisc, ci.NonHCPCluster, func() {

			By("Adding additional custom property to the existing cluster")
			updatedCustomProperties := constants.CustomProperties
			updatedCustomProperties["second_custom_property"] = "test2"

			// Apply updated custom properties to the cluster
			clusterArgs.CustomProperties = helper.StringMapPointer(updatedCustomProperties)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).ShouldNot(HaveOccurred())

			// Validating cluster's custom property update
			clusterDetails, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(clusterDetails.Body().Properties()["second_custom_property"]).Should(Equal(updatedCustomProperties["second_custom_property"]))

			By("Applying reserved property to existing cluster should not be allowed")
			updatedCustomProperties = map[string]string{
				"rosa_tf_version": "true",
			}
			clusterArgs.CustomProperties = helper.StringMapPointer(updatedCustomProperties)
			_, err = clusterService.Apply(clusterArgs)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("Can not override reserved properties keys"))
		})

	It("can edit/delete cluster properties - [id:72451]", ci.Day2, ci.Medium, ci.NonClassicCluster, ci.FeatureClusterMisc, func() {
		updatedCustomProperties := helper.CopyStringMap(originalCustomProperties)

		By("Add properties to cluster")
		updatedCustomProperties["some"] = "thing"
		updatedCustomProperties["nothing"] = ""
		clusterArgs.CustomProperties = helper.StringMapPointer(updatedCustomProperties)
		_, err = clusterService.Apply(clusterArgs)
		Expect(err).ShouldNot(HaveOccurred())

		By("Verify new properties from cluster")
		clusterDetails, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(clusterDetails.Body().Properties()["some"]).Should(Equal(updatedCustomProperties["some"]))
		Expect(clusterDetails.Body().Properties()["nothing"]).Should(Equal(updatedCustomProperties["nothing"]))

		By("Update properties to cluster")
		updatedCustomProperties["some"] = "thing2"
		clusterArgs.CustomProperties = helper.StringMapPointer(updatedCustomProperties)
		_, err = clusterService.Apply(clusterArgs)
		Expect(err).ShouldNot(HaveOccurred())

		By("Verify updated properties from cluster")
		clusterDetails, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(clusterDetails.Body().Properties()["some"]).Should(Equal(updatedCustomProperties["some"]))
		Expect(clusterDetails.Body().Properties()["nothing"]).To(Equal(updatedCustomProperties["nothing"]))

		By("Remove properties from cluster")
		clusterArgs.CustomProperties = helper.StringMapPointer(originalCustomProperties)
		_, err = clusterService.Apply(clusterArgs)
		Expect(err).ShouldNot(HaveOccurred())

		By("Verify properties are removed from cluster")
		clusterDetails, err = cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(clusterDetails.Body().Properties()["some"]).To(BeEmpty())
		Expect(clusterDetails.Body().Properties()["nothing"]).To(BeEmpty())
	})
})
