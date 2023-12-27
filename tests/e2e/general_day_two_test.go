package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("TF day2 scenrios", func() {

	Context("Author:smiron-Medium-OCP-67570 @OCP-67570 @smiron", func() {
		var dnsService *exe.DnsService
		BeforeEach(func() {
			dnsService = exe.NewDnsDomainService(con.DNSDir)
		})
		AfterEach(func() {
			err := dnsService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})
		It("OCP-67570 - Create and destroy dnsdomain via terraform provider",
			ci.Day2, ci.Medium, ci.FeatureIDP, func() {

				By("Create/Apply dns-domain resource by terraform")
				dnsArgs := &exe.DnsDomainArgs{}
				err := dnsService.Create(dnsArgs)
				Expect(err).ToNot(HaveOccurred())
			})
	})
	Context("Author:smiron-Medium-OCP-68301 @OCP-68301 @smiron", func() {
		var rhcsInfoService *exe.RhcsInfoService

		BeforeEach(func() {
			rhcsInfoService = exe.NewRhcsInfoService(con.RhcsInfoDir)
		})

		It("OCP-68301 - Verify the state of the rhcs_info data source",
			ci.Day2, ci.Medium, ci.FeatureIDP, func() {

				By("Creating/Applying rhcs-info resource by terraform")
				rhcsInfoArgs := &exe.RhcsInfoArgs{}
				Expect(rhcsInfoService.Create(rhcsInfoArgs)).ShouldNot(HaveOccurred())

				By("Comparing rhcs-info state output to OCM API output")
				currentAccountInfo, err := cms.RetrieveCurrentAccount(ci.RHCSConnection)
				Expect(err).ShouldNot(HaveOccurred())

				// Address the resource's kind and name for the state command
				rhcsInfoArgs = &exe.RhcsInfoArgs{
					ResourceKind: "rhcs_info",
					ResourceName: "info",
				}
				currentResourceState, err := rhcsInfoService.ShowState(rhcsInfoArgs)
				Expect(err).ShouldNot(HaveOccurred())

				// convert given string to a map of values
				resourceStateMap, err := h.ParseStringToMap(currentResourceState)
				Expect(err).ToNot(HaveOccurred())

				// comparsion between rhcs_info source to backend api
				Expect(resourceStateMap["account_email"]).To(Equal(currentAccountInfo.Body().Email()))
				Expect(resourceStateMap["account_id"]).To(Equal(currentAccountInfo.Body().ID()))
				Expect(resourceStateMap["account_username"]).To(Equal(currentAccountInfo.Body().Username()))
				Expect(resourceStateMap["organization_id"]).To(Equal(currentAccountInfo.Body().Organization().ID()))
				Expect(resourceStateMap["organization_external_id"]).To(Equal(currentAccountInfo.Body().Organization().ExternalID()))
				Expect(resourceStateMap["organization_name"]).To(Equal(currentAccountInfo.Body().Organization().Name()))
			})
	})
	Context("Author: smiron-Medium-OCP-64907 @OCP-64907 @smiron", func() {
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
			clusterService, err = exe.NewClusterService(profile.ManifestsDIR)
			Expect(err).ShouldNot(HaveOccurred())

			// Read terraform.tfvars file and get its content as a map
			terraformTFVarsPath := filepath.Join(profile.ManifestsDIR, "terraform.tfvars")
			terraformTFVarsContent := exe.ReadTerraformTFVars(terraformTFVarsPath)
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

		It("should validate custom property operations on cluster - day 2",
			ci.Day2, ci.Medium, ci.FeatureIDP, func() {

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
})
