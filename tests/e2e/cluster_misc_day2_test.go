package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Cluster miscellaneous", func() {
	var (
		clusterService           *exe.ClusterService
		err                      error
		profile                  *ci.Profile
		originalCustomProperties map[string]string
	)

	BeforeEach(func() {
		// Load profile from YAML file based on environment
		profile = ci.LoadProfileYamlFileByENV()

		// Initialize the cluster service
		clusterService, err = exe.NewClusterService(profile.GetClusterManifestsDir())
		Expect(err).ShouldNot(HaveOccurred())

		// Read terraform.tfvars file and get its content as a map
		terraformTFVarsContent := exe.ReadTerraformTFVars(profile.GetClusterManifestsDir())
		Expect(err).ShouldNot(HaveOccurred())

		// Capture the original custom properties
		originalCustomProperties, err = h.ParseStringToMap(terraformTFVarsContent["custom_properties"])
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		// Cleanup
		clusterArgs := &exe.ClusterCreationArgs{
			AWSRegion:        profile.Region,
			CustomProperties: originalCustomProperties,
		}

		// Restore cluster state
		err = clusterService.Apply(clusterArgs, false, true)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should validate custom property operations on cluster - [id:64907]",
		ci.Day2, ci.Medium, ci.FeatureIDP, ci.NonHCPCluster, func() {

			By("Adding additional custom property to the existing cluster")
			updatedCustomProperties := con.CustomProperties
			updatedCustomProperties["second_custom_property"] = "test2"

			// Apply updated custom properties to the cluster
			clusterArgs := &exe.ClusterCreationArgs{
				AWSRegion:        profile.Region,
				CustomProperties: updatedCustomProperties,
			}

			err = clusterService.Apply(clusterArgs, false, true)
			Expect(err).ShouldNot(HaveOccurred())

			// Validating cluster's custom property update
			clusterDetails, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(clusterDetails.Body().Properties()["second_custom_property"]).Should(Equal(updatedCustomProperties["second_custom_property"]))

			By("Applying reserved property to existing cluster should not be allowed")
			updatedCustomProperties = map[string]string{
				"rosa_tf_version": "true",
			}

			clusterArgs = &exe.ClusterCreationArgs{
				AWSRegion:        profile.Region,
				CustomProperties: updatedCustomProperties,
			}

			err = clusterService.Apply(clusterArgs, false, false)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("Can not override reserved properties keys"))
		})
})
